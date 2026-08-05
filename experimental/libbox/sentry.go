//go:build with_karing

package libbox

import (
	"os"
	"strings"
)

type SentryBoxServiceLaunchCallbackFunc func()
type SentryInitCallbackFunc func(configPath string) ([]byte, error)
type SentryCaptureMessageCallbackFunc func(panicErr error, attachMessage string) bool
type SentryCaptureExceptionCallbackFunc func(panicErrMessage string, attachMessage string, stack string) bool

var (
	SentryBoxServiceLaunchCallback    SentryBoxServiceLaunchCallbackFunc
	SentryInitCallback                SentryInitCallbackFunc
	SentryCaptureMessageCallback      SentryCaptureMessageCallbackFunc
	SentryCapturePanicMessageCallback SentryCaptureExceptionCallbackFunc
	SentryDsn                         string
	SentryDid                         string
	SentryRelease                     string
)

func SentryGetDsn() string {
	return SentryDsn
}

func SentryGetDid() string {
	return SentryDid
}

func SentryGetRelease() string {
	return SentryRelease
}

func SentryBoxServiceLaunch() {
	if SentryBoxServiceLaunchCallback != nil {
		SentryBoxServiceLaunchCallback()
	}
}

func SentryInit(configPath string) ([]byte, error) {
	if SentryInitCallback == nil {
		return os.ReadFile(configPath)
	}
	return SentryInitCallback(configPath)
}

func SentryCaptureMessage(panicErr error, attachMessage string) {
	if SentryCaptureMessageCallback == nil {
		return
	}
	SentryCaptureMessageCallback(panicErr, attachMessage)
}

func SentryCapturePanicMessage(panicErrMessage string, attachMessage string, stack string) {
	if SentryCapturePanicMessageCallback == nil {
		return
	}
	SentryCapturePanicMessageCallback(panicErrMessage, attachMessage, stack)
}

func SentryTrim(stack string) string {
	lines := strings.Split(stack, "\n")
	if len(lines) > 3 {
		if lines[1] == "runtime/debug.Stack()" {
			copy(lines[0:], lines[3:])
			lines = lines[:len(lines)-3]
		}
	}
	return strings.Join(lines, "\n")
}
