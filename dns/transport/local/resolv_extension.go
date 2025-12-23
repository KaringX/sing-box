// karing
package local

import (
	"context"
)

func GetSystemDNSConfig(ctx context.Context) *dnsConfig {
	return getSystemDNSConfig(ctx)
}
