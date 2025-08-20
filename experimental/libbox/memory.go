package libbox

import (
	"math"
	"runtime"
	runtimeDebug "runtime/debug"

	"github.com/sagernet/sing-box/common/conntrack"
)

func SetMemoryLimit(enabled bool) {
	const memoryLimit = 45 * 1024 * 1024
	const memoryLimitGo = memoryLimit / 1.5
	if enabled {
		runtimeDebug.SetGCPercent(10)
		runtimeDebug.SetMemoryLimit(memoryLimitGo)
		conntrack.KillerEnabled = true
		conntrack.MemoryLimit = memoryLimit
	} else {
		runtimeDebug.SetGCPercent(100)
		runtimeDebug.SetMemoryLimit(math.MaxInt64)
		conntrack.KillerEnabled = false
	}
}

func Gc() { //karing
	runtime.GC()
	runtimeDebug.FreeOSMemory()
}
