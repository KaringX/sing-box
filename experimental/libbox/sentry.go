// karing
package libbox

import "os"

type SentryInitCallbackFunc func(configPath string) ([]byte, error)
type SentryCaptureExceptionCallbackFunc func(exception error)

var (
	SentryInitCallback SentryInitCallbackFunc
	SentryCaptureExceptionCallback SentryCaptureExceptionCallbackFunc
	SentryDsn          string
	SentryDid          string
	SentryRelease      string
)

type SentryPanicError struct {
	Err   string
}

func (e *SentryPanicError) Error() string {
	return e.Err
}


func SentryGetDsn() string {
	return SentryDsn
}

func SentryGetDid() string {
	return SentryDid
}

func SentryGetRelease() string {
	return SentryRelease
}

func SentryInit(configPath string) ([]byte, error){
	if SentryInitCallback == nil {
		return os.ReadFile(configPath)
	}
	return SentryInitCallback(configPath)
}
func SentryCaptureException(exception error){
	if SentryCaptureExceptionCallback == nil {
		return  
	}
	SentryCaptureExceptionCallback(exception)
}
