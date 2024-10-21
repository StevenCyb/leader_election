package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_Timer_Trigger(t *testing.T) {
	t.Parallel()

	myTimer := &Timer{}
	called := false
	now := time.Now()
	myTimer.Set(time.Second)

	myTimer.OnTrigger(func() {
		require.False(t, called, "trigger should only be called once")
		called = true
		require.WithinDuration(t, now.Add(time.Second), time.Now(), 100*time.Millisecond)
	})

	myTimer.Reset()
	time.Sleep(3 * time.Second)

	require.True(t, called, "trigger should have been called")
}

func Test_Timer_Rest(t *testing.T) {
	t.Parallel()

	myTimer := &Timer{}
	myTimer.Set(time.Second)

	myTimer.OnTrigger(func() {
		require.Fail(t, "trigger should not be called")
	})

	for i := 0; i < 5; i++ {
		myTimer.Reset()
		time.Sleep(500 * time.Millisecond)
	}
}

func Test_Timer_Stop(t *testing.T) {
	t.Parallel()

	myTimer := &Timer{}
	myTimer.Set(time.Second)

	myTimer.OnTrigger(func() {
		require.Fail(t, "trigger should not be called")
	})

	myTimer.Reset()
	myTimer.Stop()
	time.Sleep(2 * time.Second)
}

func Test_Timer_Trigger_Multiple(t *testing.T) {
	t.Parallel()

	myTimer := &Timer{}
	called := 0
	now := time.Now()
	myTimer.Set(time.Second)

	myTimer.OnTrigger(func() {
		called++
		require.WithinDuration(t, now.Add(time.Second), time.Now(), 250*time.Millisecond)
		now = time.Now()
	})

	myTimer.Reset()
	time.Sleep(1200 * time.Millisecond)
	myTimer.Reset()
	time.Sleep(1200 * time.Millisecond)
	time.Sleep(300 * time.Millisecond)

	require.Equal(t, 2, called, "trigger should have been called twice")
}

func Test_Timer_Restart(t *testing.T) {
	t.Parallel()

	myTimer := &Timer{}
	called := false
	now := time.Now()
	myTimer.Set(time.Second)

	myTimer.OnTrigger(func() {
		require.False(t, called, "trigger should only be called once")
		called = true
		require.WithinDuration(t, now.Add(time.Second), time.Now(), 100*time.Millisecond)
	})

	myTimer.Reset()
	myTimer.Stop()
	myTimer.Reset()
	time.Sleep(3 * time.Second)

	require.True(t, called, "trigger should have been called")
}
