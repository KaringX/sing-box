package congestion

import (
	"time"

	"github.com/sagernet/quic-go"
	"github.com/sagernet/quic-go/congestion"
	congestion_meta1 "github.com/sagernet/sing-quic/congestion_meta1"
	congestion_meta2 "github.com/sagernet/sing-quic/congestion_meta2"
	E "github.com/sagernet/sing/common/exceptions"
)

func NewCongestionControl(name string, cwnd int, timeFunc func() time.Time) (func(conn *quic.Conn) congestion.CongestionControl, error) {
	if timeFunc == nil {
		timeFunc = time.Now
	}
	if cwnd == 0 {
		cwnd = 32
	}
	switch name {
	case "", "bbr":
		return func(conn *quic.Conn) congestion.CongestionControl {
			return congestion_meta2.NewBbrSenderWithProfile(conn.InitialPacketSize(), congestion_meta2.ProfileStandard)
		}, nil
	case "cubic":
		return func(conn *quic.Conn) congestion.CongestionControl {
			return congestion_meta1.NewCubicSender(
				congestion_meta1.DefaultClock{TimeFunc: timeFunc},
				conn.InitialPacketSize(),
				false,
			)
		}, nil
	case "reno":
		return func(conn *quic.Conn) congestion.CongestionControl {
			return congestion_meta1.NewCubicSender(
				congestion_meta1.DefaultClock{TimeFunc: timeFunc},
				conn.InitialPacketSize(),
				true,
			)
		}, nil
	default:
		return nil, E.New("unknown congestion control: ", name)
	}
}
