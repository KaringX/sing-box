//karing

package route

import (
	"context"
	"runtime"
	runtimeDebug "runtime/debug"
	"strings"

	"github.com/sagernet/sing-box/adapter"
)

func (r *NetworkManager) ResetOutboundNetwork(ctx context.Context, tags []string) {
	if r.outbound != nil {
		for _, outbound := range r.outbound.Outbounds() {
			listener, isListener := outbound.(adapter.InterfaceUpdateListener)
			if isListener {
				if strings.Contains(outbound.Tag(), outbound.Tag()) {
					listener.InterfaceUpdated(ctx)
				}
			}
		}
	}
	go func() {
		runtime.GC()
		runtimeDebug.FreeOSMemory()
	}()
}
