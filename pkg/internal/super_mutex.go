package internal

import (
	"fmt"
	"sync/atomic"
)

type SuperMutex struct {
	ch       chan struct{}
	isLocked bool
}

func NewMyMutex() *SuperMutex {
	return &SuperMutex{ch: make(chan struct{}, 1)}
}

func (m *SuperMutex) Lock() {
	m.ch <- struct{}{}
	m.isLocked = true
}

func (m *SuperMutex) TryLock() bool {
	select {
	case m.ch <- struct{}{}:
		m.isLocked = true
		return true
	default:
		return false
	}
}

func (m *SuperMutex) Unlock() {
	if !m.isLocked {
		panic("unlock of unlocked mutex")
	}

	<-m.ch
	m.isLocked = false
}

func (m *SuperMutex) IsLocked() bool {
	return m.isLocked
}

func (m *SuperMutex) String() string {
	return fmt.Sprintf("SuperMutex{isLocked: %v}", m.isLocked)
}

type SuperRWMutex struct {
	ch            chan struct{}
	readLockCount int32
	writeLock     bool
}

func NewSuperRWMutex() *SuperRWMutex {
	return &SuperRWMutex{ch: make(chan struct{}, 1)}
}

func (m *SuperRWMutex) Lock() {
	// Wait until there are no active read locks
	for atomic.LoadInt32(&m.readLockCount) > 0 {
		// Busy wait or use a more sophisticated waiting mechanism
	}
	m.ch <- struct{}{}
	m.writeLock = true
}

func (m *SuperRWMutex) Unlock() {
	if !m.writeLock {
		panic("unlock of unlocked mutex")
	}
	m.writeLock = false
	<-m.ch
}

func (m *SuperRWMutex) RLock() {
	atomic.AddInt32(&m.readLockCount, 1)
}

func (m *SuperRWMutex) RUnlock() {
	if atomic.LoadInt32(&m.readLockCount) == 0 {
		panic("unlock of unlocked read lock")
	}
	atomic.AddInt32(&m.readLockCount, -1)
}

func (m *SuperRWMutex) TryLock() bool {
	if atomic.LoadInt32(&m.readLockCount) > 0 {
		return false
	}
	select {
	case m.ch <- struct{}{}:
		m.writeLock = true
		return true
	default:
		return false
	}
}

func (m *SuperRWMutex) TryRLock() bool {
	atomic.AddInt32(&m.readLockCount, 1)
	return true
}

func (m *SuperRWMutex) IsWriteLocked() bool {
	return m.writeLock
}

func (m *SuperRWMutex) IsReadLocked() bool {
	return atomic.LoadInt32(&m.readLockCount) > 0
}

func (m *SuperRWMutex) GetReadLockCount() int32 {
	return atomic.LoadInt32(&m.readLockCount)
}

func (m *SuperRWMutex) String() string {
	return fmt.Sprintf("SuperRWMutex{writeLock: %v, readLockCount: %d}", m.writeLock, m.readLockCount)
}
