// karing
package tun

import (
	tun "github.com/sagernet/sing-tun"
)

var tunIfSuspended tun.Tun

func (t *Inbound) Suspend() {
	tunIfSuspended = t.tunIf
}

func (t *Inbound) IsSuspended() bool {
	return tunIfSuspended != nil
}

func (t *Inbound) get() tun.Tun {
	return tunIfSuspended
}

func (t *Inbound) takeAndResume() tun.Tun {
	tunInterface := tunIfSuspended
	tunIfSuspended = nil
	return tunInterface
}

func (t *Inbound) Close() {
	tunInterface := t.takeAndResume()
	if tunInterface != nil {
		tunInterface.Close()
	}
}
