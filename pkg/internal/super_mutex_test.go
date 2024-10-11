package internal

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuperMutex_Lock_Unlock(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	mutex := NewMyMutex()
	actual := 0
	expect := 1000

	for i := 0; i < expect; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mutex.Lock()
			actual = actual + 1
			mutex.Unlock()
		}()
	}

	wg.Wait()
	assert.Equal(t, expect, actual)
}

func TestSuperMutex_TryLock(t *testing.T) {
	t.Parallel()

	mutex := NewMyMutex()

	assert.True(t, mutex.TryLock())
	assert.True(t, mutex.IsLocked())
	assert.False(t, mutex.TryLock())
	mutex.Unlock()
	assert.True(t, mutex.TryLock())
	assert.False(t, mutex.TryLock())
}

func TestSuperMutex_IsLocked(t *testing.T) {
	t.Parallel()

	mutex := NewMyMutex()

	assert.False(t, mutex.IsLocked())
	mutex.Lock()
	assert.True(t, mutex.IsLocked())
	mutex.Unlock()
	assert.False(t, mutex.IsLocked())
}

func TestSuperRWMutex_Lock_Unlock(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup
	rwMutex := NewSuperRWMutex()
	actual := 0
	expect := 1000

	for i := 0; i < expect; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rwMutex.Lock()
			actual = actual + 1
			rwMutex.Unlock()
		}()
	}

	wg.Wait()
	assert.Equal(t, expect, actual)
}

func TestSuperRWMutex_TryLock(t *testing.T) {
	t.Parallel()

	rwMutex := NewSuperRWMutex()

	assert.True(t, rwMutex.TryLock())
	assert.True(t, rwMutex.IsWriteLocked())
	assert.False(t, rwMutex.TryLock())
	rwMutex.Unlock()
	assert.True(t, rwMutex.TryLock())
	assert.False(t, rwMutex.TryLock())
}

func TestSuperRWMutex_IsWriteLocked(t *testing.T) {
	t.Parallel()

	rwMutex := NewSuperRWMutex()

	assert.False(t, rwMutex.IsWriteLocked())
	rwMutex.Lock()
	assert.True(t, rwMutex.IsWriteLocked())
	rwMutex.Unlock()
	assert.False(t, rwMutex.IsWriteLocked())
}

func TestSuperRWMutex_ReadLock(t *testing.T) {
	t.Parallel()

	rwMutex := NewSuperRWMutex()

	rwMutex.RLock()
	rwMutex.RLock()
	assert.False(t, rwMutex.TryLock())
	rwMutex.RUnlock()
	assert.False(t, rwMutex.TryLock())
	rwMutex.RUnlock()
	assert.True(t, rwMutex.TryLock())

}

func TestSuperRWMutex_TryRLock(t *testing.T) {
	t.Parallel()

	rwMutex := NewSuperRWMutex()

	assert.True(t, rwMutex.TryRLock())
	assert.True(t, rwMutex.IsReadLocked())
	assert.True(t, rwMutex.TryRLock())
	assert.True(t, rwMutex.IsReadLocked())
	rwMutex.RUnlock()
	rwMutex.RUnlock()
	assert.False(t, rwMutex.IsReadLocked())
}

func TestSuperRWMutex_IsReadLocked(t *testing.T) {
	t.Parallel()

	rwMutex := NewSuperRWMutex()

	assert.False(t, rwMutex.IsReadLocked())
	rwMutex.RLock()
	assert.True(t, rwMutex.IsReadLocked())
	rwMutex.RUnlock()
	assert.False(t, rwMutex.IsReadLocked())
}

func TestSuperRWMutex_GetReadLockCount(t *testing.T) {
	t.Parallel()

	rwMutex := NewSuperRWMutex()

	assert.Equal(t, int32(0), rwMutex.GetReadLockCount())
	rwMutex.RLock()
	assert.Equal(t, int32(1), rwMutex.GetReadLockCount())
	rwMutex.RLock()
	assert.Equal(t, int32(2), rwMutex.GetReadLockCount())
	rwMutex.RUnlock()
	assert.Equal(t, int32(1), rwMutex.GetReadLockCount())
	rwMutex.RUnlock()
	assert.Equal(t, int32(0), rwMutex.GetReadLockCount())
}
