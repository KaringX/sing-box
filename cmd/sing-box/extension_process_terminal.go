//go:build with_karing && linux

package main

import (
	"os"
)

func terminateCurrentProcess() {
	os.Exit(1)
}
