//go:build with_karing && !linux && !windows && !darwin && !android

package libbox

import (
	"os"
)

func stderrRedirect(f *os.File) error {
	return nil
}
