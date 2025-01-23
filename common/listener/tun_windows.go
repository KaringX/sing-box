//karing
//go:build windows

package listener

import (
	tun "github.com/sagernet/sing-tun"
)
func SetTunnelType(name string) {
	tun.TunnelType = name
}