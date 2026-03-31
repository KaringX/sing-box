//karing
//go:build !windows

package tun

import (
	"github.com/sagernet/sing-box/log"
	tun "github.com/sagernet/sing-tun"
)

func SetTunnelType(name string) {
}

func newTunWithFallback(_ log.ContextLogger, options *tun.Options, _ bool) (tun.Tun, error) {
	return tun.New(*options)
}
