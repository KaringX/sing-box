// karing
package libbox

import (
	"encoding/json"
	"fmt"

	"github.com/sagernet/sing-box/experimental/clashapi"
)

var contextId int
var restartFlag bool

func SetRestart(restart bool) {
	restartFlag = restart
}

func GetRestart() bool {
	return restartFlag
}

func (w *platformInterfaceWrapper) GetAssetContent(path string) ([]byte, error) {
	return w.iif.GetAssetContent(path)
}

func (s *BoxService) GetConnections(includeConnections bool) string {
	if s.clashServer != nil {
		trafficManager := s.clashServer.(*clashapi.Server).TrafficManager()
		if trafficManager != nil {
			snapshot := trafficManager.Snapshot(includeConnections)
			data, err := json.Marshal(snapshot)
			if err != nil {
				return fmt.Sprintf("{err:%s}", err.Error())
			}
			return string(data)
		}
	}
	return "{}"
}
