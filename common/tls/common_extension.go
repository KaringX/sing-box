// karing
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
