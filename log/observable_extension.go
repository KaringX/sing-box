// karing
package log

import (
	"context"
	"path"
	"runtime"
	"strconv"

	"strings"

	"time"

	F "github.com/sagernet/sing/common/format"
)

var CaptureFatalMessageFunc func(message string)

func (l *observableLogger) log(ctx context.Context, level Level, deep int, args []any) {
	level = OverrideLevelFromContext(level, ctx)
	platformWriters := l.loadPlatformWriters()
	if level > l.level && len(platformWriters) == 0 && !l.needObservable {
		return
	}
	if l.writer == nil {
		return
	}
	_, file, line, _ := runtime.Caller(deep)
	fileTag := " " + path.Base(file) + ":" + strconv.Itoa(line) + " " + l.tag
	nowTime := time.Now()
	message := fileTag + F.ToString(args...)
	if level == LevelFatal || level == LevelPanic {
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
