// karing
package outbound

import (
	"time"
)

type OutboundGetLatestDownloadActiveConnectionFunc func(tag string) (bool, time.Time)

var (
	OutboundGetLatestDownloadActiveConnection OutboundGetLatestDownloadActiveConnectionFunc
)
