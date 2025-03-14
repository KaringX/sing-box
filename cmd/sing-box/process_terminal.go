//go:build with_karing && !windows

package main

import (
	"os"
)

func terminateCurrentProcess() {
	os.Exit(1)
}
