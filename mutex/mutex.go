package mutex

import (
	"sync"
	"sync/atomic"

	"gvisor.dev/gvisor/pkg/tmutex"
)

type TMutex struct {
	tmutex.Mutex

	initialized uint32
}

func NewTMutex() *TMutex {
	m := &TMutex{
		Mutex:       tmutex.Mutex{},
		initialized: 0x1,
	}
	m.Mutex.Init()
	return m
}

func (m *TMutex) Lock() {
	if atomic.CompareAndSwapUint32(&m.initialized, 0x0, 0x1) {
		m.Mutex.Init()
	}
	m.Mutex.Lock()
}

func (m *TMutex) TryLock() bool {
	if atomic.CompareAndSwapUint32(&m.initialized, 0x0, 0x1) {
		m.Mutex.Init()
	}
	return m.Mutex.TryLock()
}

func (m *TMutex) Unlock() {
	if atomic.CompareAndSwapUint32(&m.initialized, 0x0, 0x1) {
		m.Mutex.Init()
	}
	m.Mutex.Unlock()
}

type Mutex struct {
	sync.Mutex

	flag uint32
}

func (m *Mutex) Lock() {
	m.Mutex.Lock()

	atomic.StoreUint32(&m.flag, 0x1)
}

func (m *Mutex) TryLock() bool {
	if atomic.CompareAndSwapUint32(&m.flag, 0x0, 0x1) {
		m.Mutex.Lock()
		return true
	}
	return false
}

func (m *Mutex) Unlock() {
	atomic.StoreUint32(&m.flag, 0x0)

	m.Mutex.Unlock()
}

type RWMutex struct {
	sync.RWMutex

	flags [2]uint32
}

func (m *RWMutex) RLock() {
	m.RWMutex.RLock()

	atomic.StoreUint32(&m.flags[0], 0x1)
}

func (m *RWMutex) TryRLock() bool {
	if atomic.CompareAndSwapUint32(&m.flags[0], 0x0, 0x1) {
		m.RWMutex.RLock()
		return true
	}
	return false
}

func (m *RWMutex) RUnlock() {
	atomic.StoreUint32(&m.flags[0], 0x0)

	m.RWMutex.RUnlock()
}

func (m *RWMutex) Lock() {
	m.RWMutex.Lock()

	atomic.StoreUint32(&m.flags[1], 0x1)
}

func (m *RWMutex) TryLock() bool {
	if atomic.CompareAndSwapUint32(&m.flags[1], 0x0, 0x1) {
		m.RWMutex.Lock()
		return true
	}
	return false
}

func (m *RWMutex) Unlock() {
	atomic.StoreUint32(&m.flags[1], 0x0)

	m.RWMutex.Unlock()
}
