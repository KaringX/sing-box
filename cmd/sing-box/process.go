//go:build with_karing && (windows || linux)

package main

import (
	"os"
	"path/filepath"

	"github.com/shirou/gopsutil/v3/process"
)

func makeProcessSingleton() error {
	currentProcess, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return err
	}

	currentExe, err := getResolvedExePath(currentProcess)
	if err != nil {
		return err
	}

	processes, err := process.Processes()
	if err != nil {
		return err
	}
	for _, p := range processes {
		if p.Pid == currentProcess.Pid {
			continue
		}

		targetExe, err := getResolvedExePath(p)
		if err != nil {
			continue
		}
		if targetExe == currentExe {
			terminateProcess(p)
		}
	}
	return nil
}

func getResolvedExePath(p *process.Process) (string, error) {
	exePath, err := p.Exe()
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(exePath)
	if err != nil {
		return "", err
	}

	resolvedPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", err
	}

	return filepath.Clean(resolvedPath), nil
}

func terminateProcess(p *process.Process) error {
	err := p.Terminate()
	if err == nil {
		return err
	}

	return p.Kill()
}
