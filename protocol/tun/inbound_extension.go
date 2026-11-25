// karing
package tun

import (
	"sync"

	tun "github.com/sagernet/sing-tun"
)

var suspendedTunTransfer = struct {
	sync.Mutex
	tun                 tun.Tun
	closeFileDescriptor func()
}{}

func (t *Inbound) Suspend() {
	suspendedTunTransfer.Lock()
	suspendedTunTransfer.tun = t.tunIf
	suspendedTunTransfer.Unlock()
}

func (t *Inbound) IsSuspended() bool {
	suspendedTunTransfer.Lock()
	suspended := suspendedTunTransfer.tun != nil
	suspendedTunTransfer.Unlock()
	return suspended
}

func (t *Inbound) get() tun.Tun {
	suspendedTunTransfer.Lock()
	tunInterface := suspendedTunTransfer.tun
	suspendedTunTransfer.Unlock()
	return tunInterface
}

func (t *Inbound) takeAndResume() tun.Tun {
	suspendedTunTransfer.Lock()
	tunInterface := suspendedTunTransfer.tun
	suspendedTunTransfer.tun = nil
	suspendedTunTransfer.Unlock()
	return tunInterface
}

func (t *Inbound) Close() {
	suspendedTunTransfer.Lock()
	tunInterface := suspendedTunTransfer.tun
	closeFd := suspendedTunTransfer.closeFileDescriptor
	suspendedTunTransfer.tun = nil
	suspendedTunTransfer.closeFileDescriptor = nil
	suspendedTunTransfer.Unlock()
	if closeFd != nil {
		closeFd()
	}
	if tunInterface != nil {
		tunInterface.Close()
	}
}
