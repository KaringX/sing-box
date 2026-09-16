// karing
package trafficcontrol

import (
	"encoding/json"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing/common"
	F "github.com/sagernet/sing/common/format"
)

func GetMatchRuleChain(outboundManager adapter.OutboundManager, matchOutboundTag string) ([]string, string, string) {
	var (
		chain        []string
		next         string
		outbound     string
		outboundType string
	)
	if outboundManager == nil {
		return chain, outbound, outboundType
	}
	if len(matchOutboundTag) != 0 {
		next = matchOutboundTag
	} else {
		next = outboundManager.Default().Tag()
	}
	for {
		detour, loaded := outboundManager.Outbound(next)
		if !loaded {
			break
		}
		chain = append(chain, next)
		outbound = detour.Tag()
		outboundType = detour.Type()
		group, isGroup := detour.(adapter.OutboundGroup)
		if !isGroup || group == nil {
			break
		}
		next = group.Now()
	}
	return chain, outbound, outboundType
}

func (t TrackerMetadata) MarshalJSON() ([]byte, error) {
	var inbound string
	if t.Metadata.Inbound != "" {
		inbound = t.Metadata.InboundType + "/" + t.Metadata.Inbound
	} else {
		inbound = t.Metadata.InboundType
	}
	var domain string
	if t.Metadata.Domain != "" {
		domain = t.Metadata.Domain
	} else {
		domain = t.Metadata.Destination.Fqdn
	}
	var processPath string
	var packageName string //karing
	if t.Metadata.ProcessInfo != nil {
		if t.Metadata.ProcessInfo.ProcessPath != "" {
			processPath = t.Metadata.ProcessInfo.ProcessPath
		} else if len(t.Metadata.ProcessInfo.AndroidPackageNames) > 0 {
			packageName = t.Metadata.ProcessInfo.AndroidPackageNames[0] //karing
		}
		if processPath == "" {
			if t.Metadata.ProcessInfo.UserId != -1 {
				processPath = F.ToString(t.Metadata.ProcessInfo.UserId)
			}
		} else if t.Metadata.ProcessInfo.UserName != "" {
			processPath = F.ToString(processPath, " (", t.Metadata.ProcessInfo.UserName, ")")
		} else if t.Metadata.ProcessInfo.UserId != -1 {
			processPath = F.ToString(processPath, " (", t.Metadata.ProcessInfo.UserId, ")")
		}
	}
	var rule string
	if t.Rule != nil {
		rule = F.ToString(t.Rule, " => ", t.Rule.Action())
	} else {
		rule = "final"
	}
	return json.Marshal(map[string]any{
		"id": t.ID,
		"metadata": map[string]any{
			"network":         t.Metadata.Network,
			"type":            inbound,
			"sourceIP":        t.Metadata.Source.Addr,
			"destinationIP":   t.Metadata.Destination.Addr,
			"sourcePort":      F.ToString(t.Metadata.Source.Port),
			"destinationPort": F.ToString(t.Metadata.Destination.Port),
			"host":            domain,
			"dnsMode":         "normal",
			"processPath":     processPath,
			"packageName":     packageName, //karing
			"user":            t.User,      //karing
			"protocol":        t.Protocol,  //karing
		},
		"upload":      t.Upload.Load(),
		"download":    t.Download.Load(),
		"start":       t.CreatedAt,
		"chains":      t.Chain,
		"rule":        rule,
		"rulePayload": "",
	})
}

type Snapshot struct {
	SnapshotExtension
	Download    int64
	Upload      int64
	Connections []Tracker
	Memory      uint64
	MemoryTotal uint64
}

func (s *Snapshot) SnapshotMap() map[string]interface{} {
	return map[string]interface{}{
		"downloadTotal":       s.Download,
		"uploadTotal":         s.Upload,
		"connections":         common.Map(s.Connections, func(t Tracker) TrackerMetadata { return *t.Metadata() }),
		"memory":              s.Memory,
		"memoryTotal":         s.MemoryTotal,
		"startTime":           s.StartTime,
		"downloadTotalDirect": s.DownloadDirect,
		"uploadTotalDirect":   s.UploadDirect,
		"downloadSpeed":       s.DownloadSpeed,
		"uploadSpeed":         s.UploadSpeed,
		"connectionsOut":      s.ConnectionsOut,
		"connectionsOutCount": s.ConnectionsOutCount,
		"connectionsInCount":  s.ConnectionsInCount,
		"goroutines":          s.Goroutines,
		"threadCount":         s.ThreadCount,
	}
}

func (s *Snapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"downloadTotal":       s.Download,
		"uploadTotal":         s.Upload,
		"connections":         common.Map(s.Connections, func(t Tracker) TrackerMetadata { return *t.Metadata() }),
		"memory":              s.Memory,
		"memoryTotal":         s.MemoryTotal,
		"startTime":           s.StartTime,
		"downloadTotalDirect": s.DownloadDirect,
		"uploadTotalDirect":   s.UploadDirect,
		"downloadSpeed":       s.DownloadSpeed,
		"uploadSpeed":         s.UploadSpeed,
		"connectionsOut":      s.ConnectionsOut,
		"connectionsOutCount": s.ConnectionsOutCount,
		"connectionsInCount":  s.ConnectionsInCount,
		"goroutines":          s.Goroutines,
		"threadCount":         s.ThreadCount,
	})
}
