// karing
package gofree

import (
	"runtime"
	"runtime/pprof"
	"time"
)

var (
	enableFreeIdleThread bool = false
	ticker               *time.Ticker
)

func init() {
	ticker = time.NewTicker(time.Second * 5)
	go func() {
		for {
			<-ticker.C
			FreeIdleThread()
		}
	}()
}

func SetEnableFreeIdleThread(en bool) {
	enableFreeIdleThread = en
}

func FreeIdleThread() {
	if !enableFreeIdleThread {
		return
	}
	goroutineNum := int32(runtime.NumGoroutine())
	threadNum := int32(ThreadNum())

	for i := threadNum; i > goroutineNum+8; i-- {
		if i <= 32 {
			return
		}
		go func() {
			runtime.LockOSThread()
		}()
	}
}

func ThreadNum() int {
	return pprof.Lookup("threadcreate").Count()
}

//runtime.GC
//debug.FreeOSMemory
