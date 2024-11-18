package taskmonitor

import (
	"runtime"
	"time"

	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/logger"
)

type Monitor struct {
	logger  logger.Logger
	timeout time.Duration
	timer   *time.Timer
	taskName string
}
var CurrentMonitorBlockPoint string //karing
func New(logger logger.Logger, timeout time.Duration) *Monitor {
	return &Monitor{
		logger:  logger,
		timeout: timeout,
	}
}

func (m *Monitor) Start(taskName ...any) {
	m.taskName = F.ToString(taskName...) //karing
	CurrentMonitorBlockPoint = m.taskName //karing
	m.logger.Info(m.taskName, ", memory:", m.memory()) //karing
	m.timer = time.AfterFunc(m.timeout, func() {
		m.logger.Warn(F.ToString(taskName...), " take too much time to finish!")
	})
}

func (m *Monitor) Finish() {
	m.logger.Info(m.taskName, " done, memory:", m.memory()) //karing
	CurrentMonitorBlockPoint = "" //karing
	m.timer.Stop()
}
func (m *Monitor) memory() uint64{ //karing
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memory := memStats.StackInuse + memStats.HeapInuse + memStats.HeapIdle - memStats.HeapReleased
	return memory
}