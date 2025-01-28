// karing
package libbox

import "os"

type SentryInitCallbackFunc func(configPath string) ([]byte, error)

var (
	SentryInitCallback SentryInitCallbackFunc
	SentryDsn          string
	SentryDid          string
	SentryRelease      string
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

func SentryInit(configPath string) ([]byte, error){
	if SentryInitCallback == nil {
		return os.ReadFile(configPath)
	}
	return SentryInitCallback(configPath)
}
