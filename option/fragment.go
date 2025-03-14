package option

//hiddify
import (
	"math/rand"
)

type TLSFragmentOptions struct {
	Enabled bool   `json:"enabled,omitempty"`
	Size    string `json:"size,omitempty"`  // Fragment size in Bytes
	Sleep   string `json:"sleep,omitempty"` // Time to sleep between sending the fragments in milliseconds
}

func RandBetween(min int, max int) int {
	if max == min {
		return min
	}
	return rand.Intn(max-min+1) + min

}
