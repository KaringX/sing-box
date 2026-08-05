//go:build with_karing && linux

package libbox

import (
	"os"
	"syscall"
)

func stderrRedirect(f *os.File) error {
	return syscall.Dup3(int(f.Fd()), int(os.Stderr.Fd()), 0)
}
