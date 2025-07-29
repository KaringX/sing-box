package tls

import (
	"math/rand"
	"strings"
	"unicode"
)

func randomizeCase(s string) string { //hiddify
	var result strings.Builder
	for _, c := range s {
		if rand.Intn(2) == 0 {
			result.WriteRune(unicode.ToUpper(c))
		} else {
			result.WriteRune(unicode.ToLower(c))
		}
	}
	return result.String()
}

const (
	VersionTLS10 = 0x0301
	VersionTLS11 = 0x0302
	VersionTLS12 = 0x0303
	VersionTLS13 = 0x0304

	// Deprecated: SSLv3 is cryptographically broken, and is no longer
	// supported by this package. See golang.org/issue/32716.
	VersionSSL30 = 0x0300
)
