package main

import (
	"encoding/json"

	"github.com/sagernet/sing-box/experimental/libbox"
	"github.com/sagernet/sing-box/log"
)

type CommandClientHandler struct {
	logger log.Logger
}

func (cch *CommandClientHandler) Connected() {
	cch.logger.Debug("CONNECTED")
}

func (cch *CommandClientHandler) Disconnected(message string) {
	cch.logger.Debug("DISCONNECTED: ", message)
}

func (cch *CommandClientHandler) ClearLog() {
	cch.logger.Debug("clear log")
}

func (cch *CommandClientHandler) WriteLog(message string) {
	cch.logger.Debug("log: ", message)
}

func (cch *CommandClientHandler) WriteStatus(message *libbox.StatusMessage) {
	msg, _ := json.Marshal(
		map[string]int64{
			"connections-in":  int64(message.ConnectionsIn),
			"connections-out": int64(message.ConnectionsOut),
			"uplink":          message.Uplink,
			"downlink":        message.Downlink,
			"uplink-total":    message.UplinkTotal,
			"downlink-total":  message.DownlinkTotal,
		},
	)
	cch.logger.Debug("Memory: ", libbox.FormatBytes(message.Memory), ", Goroutines: ", message.Goroutines)
	log.Warn(msg)
}

func (cch *CommandClientHandler) WriteGroups(message libbox.OutboundGroupIterator) {
	if message == nil {
		return
	}
	groups := []*OutboundGroup{}
	for message.HasNext() {
		group := message.Next()
		items := group.GetItems()
		groupItems := []*OutboundGroupItem{}
		for items.HasNext() {
			item := items.Next()
			groupItems = append(groupItems,
				&OutboundGroupItem{
					Tag:          item.Tag,
					Type:         item.Type,
					URLTestTime:  item.URLTestTime,
					URLTestDelay: item.URLTestDelay,
				},
			)
		}
		groups = append(groups, &OutboundGroup{Tag: group.Tag, Type: group.Type, Selected: group.Selected, Items: groupItems})
	}
	response, _ := json.Marshal(groups)
	log.Warn(string(response))
}

func (cch *CommandClientHandler) InitializeClashMode(modeList libbox.StringIterator, currentMode string) {
	cch.logger.Debug("initial clash mode: ", currentMode)
}

func (cch *CommandClientHandler) UpdateClashMode(newMode string) {
	cch.logger.Debug("update clash mode: ", newMode)
}

type OutboundGroup struct {
	Tag      string               `json:"tag"`
	Type     string               `json:"type"`
	Selected string               `json:"selected"`
	Items    []*OutboundGroupItem `json:"items"`
}

type OutboundGroupItem struct {
	Tag          string `json:"tag"`
	Type         string `json:"type"`
	URLTestTime  int64  `json:"url-test-time"`
	URLTestDelay int32  `json:"url-test-delay"`
}
