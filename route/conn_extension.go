// karing
package route

import "github.com/sagernet/sing-box/adapter"

func (m *ConnectionManager) Connections() []adapter.OutboundContext {
	m.access.Lock()
	defer m.access.Unlock()
	connList := make([]adapter.OutboundContext, 0, m.connections.Len())
	for element := m.connections.Front(); element != nil; element = element.Next() {
		connList = append(connList, element.Value)
	}
	return connList
}
