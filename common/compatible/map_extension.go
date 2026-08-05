//go:build with_karing

package compatible

func (m *Map[K, V]) Clear() { //karing
	m.m.Clear()
}
