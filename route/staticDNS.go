// hiddify
package route

import (
	"fmt"
	"net/netip"
	"regexp"
	"strings"

	dns "github.com/sagernet/sing-dns"
)

type StaticDNSEntry struct {
	IPv4 []netip.Addr
	IPv6 []netip.Addr
}

func createEntries(items map[string][]string) map[string]StaticDNSEntry {
	entries := make(map[string]StaticDNSEntry)

	for domain, ips := range items {
		entry := StaticDNSEntry{}

		for _, ipString := range ips {
			ip, err := netip.ParseAddr(ipString)
			if err != nil {
				fmt.Printf("Invalid IP address for domain %s: %s\n", domain, ipString)
				continue
			}

			if ip.Is4() {
				entry.IPv4 = append(entry.IPv4, ip)
			} else {
				entry.IPv6 = append(entry.IPv6, ip)
			}
		}
		entries[domain] = entry
	}

	return entries
}

func errorIfEmpty(addrs []netip.Addr) ([]netip.Addr, error) {
	if len(addrs) == 0 {
		return addrs, fmt.Errorf("NotFound")
	}
	return addrs, nil
}
func (router *Router) lookupStaticIP(domain string, strategy uint8) ([]netip.Addr, error) {
	if staticDns, ok := router.staticDns[domain]; ok {

		switch strategy {
		case dns.DomainStrategyUseIPv4:
			return errorIfEmpty(staticDns.IPv4)

		case dns.DomainStrategyUseIPv6:

			return errorIfEmpty(staticDns.IPv6)

		case dns.DomainStrategyPreferIPv6:
			if len(staticDns.IPv6) == 0 {
				return errorIfEmpty(staticDns.IPv4)
			}
			return errorIfEmpty(append(staticDns.IPv6, staticDns.IPv4...))

		default:
			if len(staticDns.IPv4) == 0 {
				return errorIfEmpty(staticDns.IPv6)
			}
			return errorIfEmpty(append(staticDns.IPv4, staticDns.IPv6...))

		}
	} else {
		ip := getIpOfSslip(domain)
		if ip != "" {
			ipaddr, err := netip.ParseAddr(ip)
			if err != nil {
				return nil, err
			}
			return []netip.Addr{ipaddr}, nil
		}
		// if strings.Contains(domain, ",") {
		// 	entry := StaticDNSEntry{}
		// 	for _, ipString := range strings.Split(domain, ",") {
		// 		ip, err := netip.ParseAddr(ipString)
		// 		if err != nil {
		// 			fmt.Printf("Invalid IP address for domain %s: %s\n", domain, ipString)
		// 			continue
		// 		}

		// 		if ip.Is4() {
		// 			entry.IPv4 = append(entry.IPv4, ip)
		// 		} else {
		// 			entry.IPv6 = append(entry.IPv6, ip)
		// 		}
		// 	}
		// 	fmt.Println("Adding ",domain, entry)
		// 	router.staticDns[domain] = entry
		// 	return router.lookupStaticIP(domain, strategy)
		// }
		return nil, fmt.Errorf("NotFound")
	}

}

const ipv4Pattern = `((25[0-5]|2[0-4][0-9]|[0-1]?[0-9]?[0-9])[\.-](25[0-5]|2[0-4][0-9]|[0-1]?[0-9]?[0-9])[\.-](25[0-5]|2[0-4][0-9]|[0-1]?[0-9]?[0-9])[\.-](25[0-5]|2[0-4][0-9]|[0-1]?[0-9]?[0-9])).sslip.io$`
const ipv6Pattern = `((([0-9a-fA-F]{1,4}-){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}-){1,7}-|([0-9a-fA-F]{1,4}-){1,6}-[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}-){1,5}(-[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}-){1,4}(-[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}-){1,3}(-[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}-){1,2}(-[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}-((-[0-9a-fA-F]{1,4}){1,6})|-((-[0-9a-fA-F]{1,4}){1,7}|-)|fe80-(-[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|--(ffff(-0{1,4}){0,1}-){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}-){1,4}-((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))).sslip.io$`

var ipv4Regex, _ = regexp.Compile(ipv4Pattern)
var ipv6Regex, _ = regexp.Compile(ipv6Pattern)

func getIpOfSslip(sni string) string {
	if !strings.HasSuffix(sni, ".sslip.io") {
		return ""
	}
	submatches := ipv4Regex.FindStringSubmatch(sni)
	if len(submatches) > 1 {
		return strings.ReplaceAll(submatches[1], "-", ".")
	} else {
		submatches := ipv6Regex.FindStringSubmatch(sni)
		if len(submatches) > 1 {
			return strings.ReplaceAll(submatches[1], "-", ":")
		}
	}
	return ""

}
