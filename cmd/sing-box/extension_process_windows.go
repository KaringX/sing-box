//go:build with_karing && windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	E "github.com/sagernet/sing/common/exceptions"
	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/windows"
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

	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(snapshot)
	var procEntry windows.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))
	if err = windows.Process32First(snapshot, &procEntry); err != nil {
		return err
	}
	for {
		if procEntry.ProcessID != uint32(currentProcess.Pid) {
			targetExe := filepath.Clean(syscall.UTF16ToString(procEntry.ExeFile[:]))
			targetExeName := filepath.Base(targetExe)
			if strings.EqualFold(targetExeName, currentExeName) {
				process, err := process.NewProcess(int32(procEntry.ProcessID))
				if err != nil {
					return err
				}
				err = terminateProcess(process, process.Pid, targetExeName)
				if err != nil {
					return err
				}
			}
		}

		err = windows.Process32Next(snapshot, &procEntry)
		if err != nil {
			break
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
