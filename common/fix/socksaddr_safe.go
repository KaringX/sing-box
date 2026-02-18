// karing
package metadata

import (
	"net"
	"net/netip"
	"strconv"

	M "github.com/sagernet/sing/common/metadata"
)

// SafeSocksaddrString safely converts Socksaddr to string with port, recovering from potential panics
// caused by invalid unique.Handle in netip.Addr.Zone() (Go 1.23+).
// This is a global solution to prevent SIGSEGV crashes throughout the codebase.
func SafeSocksaddrString(addr M.Socksaddr) string {
	var result string
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Fallback: manually construct address string without problematic zone
				result = safeAddrPortString(addr.Addr, addr.Port)
			}
		}()
		result = addr.String()
	}()
	return result
}

// SafeSocksaddrAddrString safely converts Socksaddr to IP address string without port
func SafeSocksaddrAddrString(addr M.Socksaddr) string {
	var result string
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Fallback: manually construct address string without problematic zone
				result = safeAddrString(addr.Addr)
			}
		}()
		result = addr.AddrString()
	}()
	return result
}

// SafeAddrPortString safely converts netip.AddrPort to string
func SafeAddrPortString(addrPort netip.AddrPort) string {
	var result string
	func() {
		defer func() {
			if r := recover(); r != nil {
				result = safeAddrPortString(addrPort.Addr(), addrPort.Port())
			}
		}()
		result = addrPort.String()
	}()
	return result
}

// SafeAddrString safely converts netip.Addr to string
func SafeAddrString(addr netip.Addr) string {
	var result string
	func() {
		defer func() {
			if r := recover(); r != nil {
				result = safeAddrString(addr)
			}
		}()
		result = addr.String()
	}()
	return result
}

// safeAddrPortString constructs address:port string without accessing potentially invalid zone
func safeAddrPortString(addr netip.Addr, port uint16) string {
	addrStr := safeAddrString(addr)
	return net.JoinHostPort(addrStr, strconv.Itoa(int(port)))
}

// safeAddrString constructs IP address string without accessing potentially invalid zone
func safeAddrString(addr netip.Addr) string {
	if !addr.IsValid() {
		return "0.0.0.0"
	}

	// For IPv6 addresses, strip the zone to avoid accessing invalid unique.Handle
	// The zone information is nice-to-have but not critical for most use cases
	if addr.Is6() {
		// Create a new address without zone
		addr = addr.WithZone("")
	}

	// Now it's safe to call String() as there's no invalid zone handle
	return addr.String()
}
