package log

import (
	"context"
	"io"
	"os"

	"path"
	"runtime"
	"strconv"
	"strings"

	"sync"
	"sync/atomic"

	"time"

	"github.com/sagernet/sing/common"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/observable"
	"gopkg.in/natefinch/lumberjack.v2"
)

var _ ObservableFactory = (*defaultFactory)(nil)
var CaptureFatalMessageFunc func(message string) //karing

type defaultFactory struct {
	ctx               context.Context
	formatter         Formatter
	platformFormatter Formatter
	logger            *lumberjack.Logger //karing
	writer            io.Writer
	file              *os.File
	filePath          string
	platformWriters   atomic.Pointer[[]PlatformWriter]
	needObservable    bool
	level             Level
	subscriber        *observable.Subscriber[Entry]
	observer          *observable.Observer[Entry]
	startAccess       sync.Mutex
	started           atomic.Bool
	pendingEntries    []pendingEntry
}

type pendingEntry struct {
	ctx       context.Context
	level     Level
	tag       string
	message   string
	timestamp time.Time
}

func NewDefaultFactory(
	ctx context.Context,
	formatter Formatter,
	writer io.Writer,
	filePath string,
	platformWriter PlatformWriter,
	needObservable bool,
) ObservableFactory {
	factory := &defaultFactory{
		ctx:       ctx,
		formatter: formatter,
		platformFormatter: Formatter{
			BaseTime:         formatter.BaseTime,
			DisableLineBreak: true,
		},
		writer:         writer,
		filePath:       filePath,
		needObservable: needObservable,
		level:          LevelTrace,
		subscriber:     observable.NewSubscriber[Entry](128),
	}
	if platformWriter != nil {
		factory.platformWriters.Store(&[]PlatformWriter{platformWriter})
	}
	/*if platformWriter != nil {
		factory.platformFormatter.DisableColors = platformWriter.DisableColors()
	}*/
	return factory
}

func (f *defaultFactory) Start() error {
	var err error
	if f.filePath != "" {
		f.logger = &lumberjack.Logger{ //karing
			Filename:   f.filePath,
			MaxSize:    20,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
		}
		f.writer = f.logger //karing
		/* //karing
		logFile, openErr := filemanager.OpenFile(f.ctx, f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if openErr != nil {
			err = openErr
		} else {
			f.writer = logFile
			f.file = logFile
		}
		*/
	}
	if f.needObservable {
		f.observer = observable.NewObserver[Entry](f.subscriber, 64)
	}
	f.startAccess.Lock()
	pendingEntries := f.pendingEntries
	f.pendingEntries = nil
	f.started.Store(true)
	f.startAccess.Unlock()
	for _, entry := range pendingEntries {
		f.output(entry.ctx, entry.level, entry.tag, entry.message, entry.timestamp)
	}
	return err
}

func (f *defaultFactory) Close() error {
	f.startAccess.Lock()
	f.pendingEntries = nil
	f.startAccess.Unlock()
	return common.Close(
		common.PtrOrNil(f.logger), //karing
		common.PtrOrNil(f.file),
		f.subscriber,
	)
}

func (f *defaultFactory) AttachPlatformWriter(writer PlatformWriter) {
	writers := append(f.loadPlatformWriters(), writer)
	f.platformWriters.Store(&writers)
}

func (f *defaultFactory) loadPlatformWriters() []PlatformWriter {
	writers := f.platformWriters.Load()
	if writers == nil {
		return nil
	}
	return *writers
}

func (f *defaultFactory) Level() Level {
	return f.level
}

func (f *defaultFactory) SetLevel(level Level) {
	f.level = level
}

func (f *defaultFactory) Logger() ContextLogger {
	return f.NewLogger("")
}

func (f *defaultFactory) NewLogger(tag string) ContextLogger {
	return &observableLogger{f, tag}
}

func (f *defaultFactory) Subscribe() (subscription observable.Subscription[Entry], done <-chan struct{}, err error) {
	return f.observer.Subscribe()
}

func (f *defaultFactory) UnSubscribe(sub observable.Subscription[Entry]) {
	f.observer.UnSubscribe(sub)
}

func (f *defaultFactory) output(ctx context.Context, level Level, tag string, message string, timestamp time.Time) {
	contextId, ok := f.ctx.Value(CtxKeyLogContextIdName).(string) // karing
	if !ok {                                                      //karing
		contextId = ""
	}
	if f.needObservable {
		formatted, formattedSimple := f.formatter.FormatWithSimple(ctx, contextId, level, tag, message, timestamp) // karing
		if level <= f.level {
			if level == LevelPanic {
				panic(formatted)
			}
			f.writer.Write([]byte(formatted))
			if level == LevelFatal {
				os.Exit(1)
			}
		}
		f.subscriber.Emit(Entry{level, formattedSimple})
	} else if level <= f.level {
		formatted := f.formatter.Format(ctx, contextId, level, tag, message, timestamp)
		if level == LevelPanic {
			panic(formatted)
		}
		f.writer.Write([]byte(formatted))
		if level == LevelFatal {
			os.Exit(1)
		}
	}
	platformWriters := f.loadPlatformWriters()
	if len(platformWriters) > 0 {
		platformMessage := f.platformFormatter.Format(ctx, contextId, level, tag, message, timestamp)
		for _, platformWriter := range platformWriters {
			platformWriter.WriteMessage(level, platformMessage)
		}
	}
}

var _ ContextLogger = (*observableLogger)(nil)

type observableLogger struct {
	*defaultFactory
	tag string
}

// karing
func (l *observableLogger) log(ctx context.Context, level Level, deep int, args []any) {
	level = OverrideLevelFromContext(level, ctx)
	platformWriters := l.loadPlatformWriters()
	if level > l.level && len(platformWriters) == 0 && !l.needObservable {
		return
	}
	if l.writer == nil { //karing
		return
	}
	contextId, ok := l.ctx.Value(CtxKeyLogContextIdName).(string) // karing
	if !ok {                                                      //karing
		contextId = ""
	}
	_, file, line, _ := runtime.Caller(deep)                              // karing
	tag := " " + path.Base(file) + ":" + strconv.Itoa(line) + " " + l.tag // karing
	nowTime := time.Now()
	//message := F.ToString(args...)//karing
	message := l.formatter.Format(ctx, contextId, level, tag, F.ToString(args...), nowTime) //karing
	if level == LevelFatal || level == LevelPanic {                                         //karing
		if CaptureFatalMessageFunc != nil {
			index := strings.Index(message, "FATAL")
			if index >= 0 {
				CaptureFatalMessageFunc(message[index:])
			} else {
				CaptureFatalMessageFunc(message)
			}
		}
	}
	if !l.started.Load() && level != LevelFatal && level != LevelPanic {
		l.startAccess.Lock()
		if !l.started.Load() {
			l.pendingEntries = append(l.pendingEntries, pendingEntry{ctx, level, l.tag, message, nowTime})
			l.startAccess.Unlock()
			return
		}
		l.startAccess.Unlock()
	}
	l.output(ctx, level, l.tag, message, nowTime)
}

func (l *observableLogger) Trace(args ...any) {
	l.log(context.Background(), LevelTrace, 2, args) // karing
}

func (l *observableLogger) Debug(args ...any) {
	l.log(context.Background(), LevelDebug, 2, args) // karing
}

func (l *observableLogger) Info(args ...any) {
	l.log(context.Background(), LevelInfo, 2, args) // karing
}

func (l *observableLogger) Warn(args ...any) {
	l.log(context.Background(), LevelWarn, 2, args) // karing
}

func (l *observableLogger) Error(args ...any) {
	l.log(context.Background(), LevelError, 2, args) // karing
}

func (l *observableLogger) Fatal(args ...any) {
	l.log(context.Background(), LevelFatal, 2, args) // karing
}

func (l *observableLogger) Panic(args ...any) {
	l.log(context.Background(), LevelPanic, 2, args) // karing
}

func (l *observableLogger) TraceContext(ctx context.Context, args ...any) {
	deep, ok := ctx.Value(CtxKeyLogContextStackDeepName).(int) // karing
	if !ok {                                                   //karing
		deep = 2
	}
	l.log(ctx, LevelTrace, deep, args) // karing
}

func (l *observableLogger) DebugContext(ctx context.Context, args ...any) {
	deep, ok := ctx.Value(CtxKeyLogContextStackDeepName).(int) // karing
	if !ok {                                                   //karing
		deep = 2
	}
	l.log(ctx, LevelDebug, deep, args) // karing
}

func (l *observableLogger) InfoContext(ctx context.Context, args ...any) {
	deep, ok := ctx.Value(CtxKeyLogContextStackDeepName).(int) // karing
	if !ok {                                                   //karing
		deep = 2
	}
	l.log(ctx, LevelInfo, deep, args) // karing
}

func (l *observableLogger) WarnContext(ctx context.Context, args ...any) {
	deep, ok := ctx.Value(CtxKeyLogContextStackDeepName).(int) // karing
	if !ok {                                                   //karing
		deep = 2
	}
	l.log(ctx, LevelWarn, deep, args) // karing
}

func (l *observableLogger) ErrorContext(ctx context.Context, args ...any) {
	deep, ok := ctx.Value(CtxKeyLogContextStackDeepName).(int) // karing
	if !ok {                                                   //karing
		deep = 2
	}
	l.log(ctx, LevelError, deep, args) // karing
}

func (l *observableLogger) FatalContext(ctx context.Context, args ...any) {
	deep, ok := ctx.Value(CtxKeyLogContextStackDeepName).(int) // karing
	if !ok {                                                   //karing
		deep = 2
	}
	l.log(ctx, LevelFatal, deep, args) // karing
}

func (l *observableLogger) PanicContext(ctx context.Context, args ...any) {
	deep, ok := ctx.Value(CtxKeyLogContextStackDeepName).(int) // karing
	if !ok {                                                   //karing
		deep = 2
	}
	l.log(ctx, LevelPanic, deep, args) // karing
}
