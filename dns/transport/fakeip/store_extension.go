// karing
package fakeip

func (s *Store) Suspend() {
	if s.storage == nil {
		return
	}
	if memoryStorage, isMemoryStorage := s.storage.(*SharedMemoryStorage); isMemoryStorage {
		memoryStorage.Suspend()
	}
}

func (s *Store) Resume() {
	if s.storage == nil {
		return
	}
	if memoryStorage, isMemoryStorage := s.storage.(*SharedMemoryStorage); isMemoryStorage {
		memoryStorage.Resume()
	}
}
