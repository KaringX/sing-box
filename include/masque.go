//go:build with_masque

// https://github.com/shtorm-7/sing-box-extended
package include

import (
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/protocol/masque"
)

func registerMASQUEOutbound(registry *outbound.Registry) {
	masque.RegisterOutbound(registry)
}
