package log

import (
	"context"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing/common"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/observable"
	"gopkg.in/natefinch/lumberjack.v2"
)

var _ Factory = (*defaultFactory)(nil)

type defaultFactory struct {
	ctx               context.Context
	formatter         Formatter
	platformFormatter Formatter
	logger            *lumberjack.Logger //karing
	writer            io.Writer
	file              *os.File
	filePath          string
	platformWriter    PlatformWriter
	needObservable    bool
	level             Level
	subscriber        *observable.Subscriber[Entry]
	observer          *observable.Observer[Entry]
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
		platformWriter: platformWriter,
		needObservable: needObservable,
		level:          LevelTrace,
		subscriber:     observable.NewSubscriber[Entry](128),
	}
	if platformWriter != nil {
		factory.platformFormatter.DisableColors = platformWriter.DisableColors()
	}
	if needObservable {
		factory.observer = observable.NewObserver[Entry](factory.subscriber, 64)
	}
	return factory
}

func (f *defaultFactory) Start() error {
	if f.filePath != "" {
		f.logger = &lumberjack.Logger{ //karing
			Filename:   f.filePath,  
			MaxSize:    50,                      
			MaxBackups: 0,                       
			MaxAge:     0,                        
			Compress:   false,                     
		}
		f.writer = f.logger //karing
		/* //karing
		logFile, err := filemanager.OpenFile(f.ctx, f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		f.writer = logFile
		f.file = logFile
		*/
	}
	return nil
}

func (f *defaultFactory) Close() error {
	return common.Close(
		f.logger, //karing
		common.PtrOrNil(f.file),
		f.subscriber,
	)
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

var _ ContextLogger = (*observableLogger)(nil)

type observableLogger struct {
	*defaultFactory
	tag string
}
// karing
func (l *observableLogger) log(ctx context.Context, level Level, deep int, args []any) {
	level = OverrideLevelFromContext(level, ctx)
	if level > l.level {
		return
	}
	if(l.writer == nil){ //karing
		return
	}
	_, file, line, _ := runtime.Caller(deep)  // karing
	tag := " " + path.Base(file) + ":" + strconv.Itoa(line) + " " + l.tag  // karing
	nowTime := time.Now()
	if l.needObservable {
		message, messageSimple := l.formatter.FormatWithSimple(ctx, level, tag, F.ToString(args...), nowTime)
		if level == LevelPanic {
			panic(message)
		}
		l.writer.Write([]byte(message))
		if level == LevelFatal {
			sentry.CaptureMessage(message) //karing
			sentry.Flush(time.Second * 3)  //karing
			log.Fatal(message)
		}
		l.subscriber.Emit(Entry{level, messageSimple})
	} else {
		message := l.formatter.Format(ctx, level, tag, F.ToString(args...), nowTime)
		if level == LevelPanic {
			panic(message)
		}
		l.writer.Write([]byte(message))
		if level == LevelFatal {
			sentry.CaptureMessage(message) //karing
			sentry.Flush(time.Second * 3)  //karing
			log.Fatal(message)
		}
	}
	if C.Build == "debug" { //karing
		if l.platformWriter != nil {
			l.platformWriter.WriteMessage(level, l.platformFormatter.Format(ctx, level, l.tag, F.ToString(args...), nowTime))
		}
	}

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
	l.log(ctx, LevelTrace, 2, args) // karing
}

func (l *observableLogger) DebugContext(ctx context.Context, args ...any) {
	l.log(ctx, LevelDebug, 2, args) // karing
}

func (l *observableLogger) InfoContext(ctx context.Context, args ...any) {
	l.log(ctx, LevelInfo, 2, args) // karing
}

func (l *observableLogger) WarnContext(ctx context.Context, args ...any) {
	l.log(ctx, LevelWarn, 2, args) // karing
}

func (l *observableLogger) ErrorContext(ctx context.Context, args ...any) {
	l.log(ctx, LevelError, 2, args) // karing
}

func (l *observableLogger) FatalContext(ctx context.Context, args ...any) {
	l.log(ctx, LevelFatal, 2, args) // karing
}

func (l *observableLogger) PanicContext(ctx context.Context, args ...any) {
	l.log(ctx, LevelPanic, 2, args) // karing
}
