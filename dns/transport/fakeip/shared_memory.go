// karing
package fakeip

import (
	"net/netip"
	"sync"
	"sync/atomic"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing/common/logger"
)

var _ adapter.FakeIPStorage = (*SharedMemoryStorage)(nil)

type SharedMemoryStorage struct {
	addressAccess sync.RWMutex
	domainAccess  sync.RWMutex
	addressCache  map[netip.Addr]string
	domainCache4  map[string]netip.Addr
	domainCache6  map[string]netip.Addr
	suspended     atomic.Bool             //karing run reuse
	metadata      *adapter.FakeIPMetadata //karing run reuse
}

var sharedMemory = &SharedMemoryStorage{
	addressCache: make(map[netip.Addr]string),
	domainCache4: make(map[string]netip.Addr),
	domainCache6: make(map[string]netip.Addr),
}

func NewSharedMemoryStorage() *SharedMemoryStorage {
	return sharedMemory
}

func (s *SharedMemoryStorage) FakeIPMetadata() *adapter.FakeIPMetadata {
	return s.metadata //karing run reuse
}

func (s *SharedMemoryStorage) FakeIPSaveMetadata(metadata *adapter.FakeIPMetadata) error {
	s.metadata = metadata //karing run reuse
	return nil
}

func (s *SharedMemoryStorage) FakeIPSaveMetadataAsync(metadata *adapter.FakeIPMetadata) {
	_ = s.FakeIPSaveMetadata(metadata) //karing run reuse
}

func (s *SharedMemoryStorage) FakeIPStore(address netip.Addr, domain string) error {
	s.addressAccess.Lock()
	s.domainAccess.Lock()
	if oldDomain, loaded := s.addressCache[address]; loaded {
		if oldDomain != domain { //karing run reuse
			log.Warn("fakeip address already associated with a different domain: ", oldDomain, " -> ", domain, " address:", address.String()) //karing
		}
		if address.Is4() {
			delete(s.domainCache4, oldDomain)
		} else {
			delete(s.domainCache6, oldDomain)
		}
	}
	s.addressCache[address] = domain
	if address.Is4() {
		s.domainCache4[domain] = address
	} else {
		s.domainCache6[domain] = address
	}
	s.domainAccess.Unlock()
	s.addressAccess.Unlock()
	return nil
}

func (s *SharedMemoryStorage) FakeIPStoreAsync(address netip.Addr, domain string, logger logger.Logger) {
	_ = s.FakeIPStore(address, domain)
}

func (s *SharedMemoryStorage) FakeIPLoad(address netip.Addr) (string, bool) {
	s.addressAccess.RLock()
	defer s.addressAccess.RUnlock()
	domain, loaded := s.addressCache[address]
	return domain, loaded
}

func (s *SharedMemoryStorage) FakeIPLoadDomain(domain string, isIPv6 bool) (netip.Addr, bool) {
	s.domainAccess.RLock()
	defer s.domainAccess.RUnlock()
	if !isIPv6 {
		address, loaded := s.domainCache4[domain]
		return address, loaded
	} else {
		address, loaded := s.domainCache6[domain]
		return address, loaded
	}
}

func (s *SharedMemoryStorage) FakeIPReset() error {
	if s.suspended.Load() { //karing run reuse
		return nil
	}
	s.addressAccess.Lock()
	s.domainAccess.Lock()
	s.addressCache = make(map[netip.Addr]string)
	s.domainCache4 = make(map[string]netip.Addr)
	s.domainCache6 = make(map[string]netip.Addr)
	s.domainAccess.Unlock()
	s.addressAccess.Unlock()
	return nil
}

func (s *SharedMemoryStorage) Suspend() {
	s.suspended.Store(true)
}

func (s *SharedMemoryStorage) Resume() {
	s.suspended.Store(false)
}
