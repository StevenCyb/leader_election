package raft

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
)

func TestRandomTime(t *testing.T) {
	minTime := 150
	maxTime := 300
	randomTime := minTime + rand.Intn(maxTime-minTime+1)

	require.Greater(t, randomTime, minTime)
	require.Less(t, randomTime, maxTime)
}
