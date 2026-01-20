// karing
package inbound

import "net/netip"

func (m *Manager) SetTunAddressPrefix(address []netip.Prefix) {
	m.tunAddress = address
}

func (m *Manager) GetTunAddressPrefix() []netip.Prefix {
	return m.tunAddress
}
