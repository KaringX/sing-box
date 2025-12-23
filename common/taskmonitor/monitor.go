package taskmonitor

import (
	"time"

	D "github.com/sagernet/sing-box/common/debug"
	F "github.com/sagernet/sing/common/format"
	"github.com/sagernet/sing/common/logger"
)

type Monitor struct {
	logger  logger.Logger
	timeout time.Duration
	timer   *time.Timer
}

func New(logger logger.Logger, timeout time.Duration) *Monitor {
	return &Monitor{
		logger:  logger,
		timeout: timeout,
	}
}

func (m *Monitor) Start(taskName ...any) {
	name := F.ToString(taskName...)          //karing
	goroutineId := D.GetCurrentGoroutineId() //karing
	m.timer = time.AfterFunc(m.timeout, func() {
		stack := D.GetGoroutineStack(goroutineId)                                   //karing
		m.logger.Warn(name, " take too much time to finish! stack:\n", stack, "\n") //karing
	})
}

func (m *Monitor) Finish() {
	m.timer.Stop()
}
