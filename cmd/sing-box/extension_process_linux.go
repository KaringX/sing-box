//go:build with_karing && linux

package main

import (
	"os"
	"path/filepath"
	"strings"

	E "github.com/sagernet/sing/common/exceptions"
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
	currentExeName := filepath.Base(currentExe)

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
		targetExeName := filepath.Base(targetExe)
		if strings.EqualFold(targetExeName, currentExeName) {
			err = terminateProcess(p, p.Pid, targetExeName)
			if err != nil {
				return err
			}
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

func terminateProcess(p *process.Process, pid int32, processName string) error {
	err := p.Terminate()
	if err == nil {
		return nil
	}

	err = p.Kill()
	if err != nil {
		return E.Cause(err, "kill process [", processName, " pid=", pid, "] failed, please try to restart your device")
	}
	return nil
}
