//go:build linux

// karing
package libbox

import (
	"os"
	"syscall"
)

func stderrRedirect(f *os.File) error {
	return syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
}