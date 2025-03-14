//go:build windows

// karing
package libbox

import (
	"os"
	"syscall"
)

 

func setStdHandle(stdhandle int32, handle syscall.Handle) error {
	var (
		kernel32	*syscall.DLL
		procSetStdHandle *syscall.Proc
		err         error
	)
	kernel32 ,err = syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	procSetStdHandle, err = kernel32.FindProc("SetStdHandle")
	if err != nil {
		return err
	}
	r0, _, e1 := syscall.Syscall(procSetStdHandle.Addr(), 2, uintptr(stdhandle), uintptr(handle), 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}

func stderrRedirect(f *os.File) error {
	return setStdHandle(syscall.STD_ERROR_HANDLE, syscall.Handle(f.Fd()))
}