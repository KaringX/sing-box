//go:build !linux && !windows && !darwin && !android

// karing
package libbox

import (
	"os"
)

func stderrRedirect(f *os.File) error {
	return nil
}