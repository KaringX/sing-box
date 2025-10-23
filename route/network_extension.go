// karing
package route

import (
	"runtime"
	runtimeDebug "runtime/debug"
	"strings"

	"github.com/sagernet/sing-box/adapter"
)

func (r *NetworkManager) ResetOutboundNetwork(tags []string) {
	if r.outbound != nil {
		for _, outbound := range r.outbound.Outbounds() {
			listener, isListener := outbound.(adapter.InterfaceUpdateListener)
			if isListener {
				if strings.Contains(outbound.Tag(), outbound.Tag()) {
					listener.InterfaceUpdated()
				}
			}
		}
	}
	go func() {
		runtime.GC()
		runtimeDebug.FreeOSMemory()
	}()
}
