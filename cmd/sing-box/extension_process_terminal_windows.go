//go:build windows

// karing
package main

import (
	"os"
	"syscall"
)

func terminateCurrentProcess() {
	handle, err := syscall.GetCurrentProcess()
	if err == nil {
		syscall.TerminateProcess(handle, uint32(1))
	} else {
		os.Exit(1)
	}
}
