//go:build go1.20

package dialer

import (
	"net"
)

type tcpDialer = ExtendedTCPDialer //hiddify

func newTCPDialer(dialer net.Dialer, tfoEnabled bool, tlsFragment *TLSFragment) (tcpDialer, error) { //hiddify
	return tcpDialer{Dialer: dialer, DisableTFO: !tfoEnabled, TLSFragment: tlsFragment}, nil //hiddify
}

func dialerFromTCPDialer(dialer tcpDialer) net.Dialer {
	return dialer.Dialer
}
