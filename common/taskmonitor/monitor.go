package taskmonitor

import (
	"time"

	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/logger"
)

type Monitor struct {
	logger   logger.Logger
	timeout  time.Duration
	timer    *time.Timer
	taskName string //karing
}

func New(logger logger.Logger, timeout time.Duration) *Monitor {
	return &Monitor{
		logger:  logger,
		timeout: timeout,
	}
}

func (m *Monitor) Start(taskName ...any) {
	m.taskName = F.ToString(taskName...) //karing
	m.timer = time.AfterFunc(m.timeout, func() {
		m.logger.Warn(F.ToString(taskName...), " take too much time to finish!")
	})
}

func (m *Monitor) Finish() {
	m.timer.Stop()
}
