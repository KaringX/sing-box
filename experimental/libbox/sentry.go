// karing
package libbox

import (
	"os"
	"strings"
)

type SentryBoxServiceLaunchCallbackFunc func()
type SentryInitCallbackFunc func(configPath string) ([]byte, error)
type SentryCaptureMessageCallbackFunc func(message error) bool
type SentryCaptureExceptionCallbackFunc func(recoverMessage string, attachMessage string, stack string) bool

var (
	SentryBoxServiceLaunchCallback SentryBoxServiceLaunchCallbackFunc
	SentryInitCallback             SentryInitCallbackFunc
	SentryCaptureMessageCallback   SentryCaptureMessageCallbackFunc
	SentryCaptureExceptionCallback SentryCaptureExceptionCallbackFunc
	SentryDsn                      string
	SentryDid                      string
	SentryRelease                  string
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

func SentryCaptureMessage(message error) {
	if SentryCaptureMessageCallback == nil {
		return
	}
	SentryCaptureMessageCallback(message)
}
func SentryCaptureException(recoverMessage string, attachMessage string, stack string) {
	if SentryCaptureExceptionCallback == nil {
		return
	}
	SentryCaptureExceptionCallback(recoverMessage, attachMessage, stack)
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
