// karing
package libbox

type SentryInitCallbackFunc func(configPath string)

var (
	SentryInitCallback SentryInitCallbackFunc
	SentryDsn          string
	SentryDid          string
	SentryVersion      string
)

func SentryGetDsn() string {
	return SentryDsn
}

func SentryGetDid() string {
	return SentryDid
}

func SentryGetVersion() string {
	return SentryVersion
}

func SentryInit(configPath string) {
	if SentryInitCallback == nil {
		return
	}
	SentryInitCallback(configPath)
}
