// karing
package libbox

import "os"

type SentryInitCallbackFunc func(configPath string) ([]byte, error)
type SentryCaptureMessageCallbackFunc func(message error)
type SentryCaptureExceptionCallbackFunc func(recoverMessage string, attachMessage string, stack string)

var (
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
