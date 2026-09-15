// karing
package dhcp

import (
	"time"

	M "github.com/sagernet/sing/common/metadata"
)

var (
	cachedServers   []M.Socksaddr
	cachedUpdatedAt time.Time
	//GetServersFromSystemDNS func(ctx context.Context) []M.Socksaddr
)
