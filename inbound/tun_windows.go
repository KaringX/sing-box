//karing
//go:build windows

package inbound

import (
	tun "github.com/sagernet/sing-tun"
)
func SetTunnelType(name string) {
	tun.TunnelType = name
}