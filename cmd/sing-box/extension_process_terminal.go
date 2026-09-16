//go:build linux

// karing
package main

import (
	"os"
)

func terminateCurrentProcess() {
	os.Exit(1)
}
