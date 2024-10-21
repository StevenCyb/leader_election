package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderedList(t *testing.T) {
	l := OrderedList[int]{}

	l.AddOrdered(5)
	l.AddOrdered(3)
	l.AddOrdered(0)
	l.AddOrdered(9)
	l.AddOrdered(1)

	require.Equal(t, OrderedList[int]{0, 1, 3, 5, 9}, l)
}

func TestOrderedList_RemoveOrdered(t *testing.T) {
	l := OrderedList[int]{10, 11, 12, 13, 14, 15, 16}

	l.RemoveOrdered(13)
	l.RemoveOrdered(10)
	l.RemoveOrdered(16)

	require.Equal(t, OrderedList[int]{11, 12, 14, 15}, l)
}

func TestOrderedList_GetIndexFor(t *testing.T) {
	l := OrderedList[int]{10, 11, 12, 13, 14, 15, 16}

	require.Equal(t, 0, l.GetIndexFor(10))
	require.Equal(t, 3, l.GetIndexFor(13))
	require.Equal(t, 6, l.GetIndexFor(16))
}

func TestOrderedList_GetValueForIndexLooped(t *testing.T) {
	l := OrderedList[int]{10, 11, 12, 13, 14, 15, 16}

	require.Equal(t, 10, l.GetValueForIndexLooped(0))
	require.Equal(t, 13, l.GetValueForIndexLooped(3))
	require.Equal(t, 16, l.GetValueForIndexLooped(6))
	require.Equal(t, 10, l.GetValueForIndexLooped(7))
	require.Equal(t, 11, l.GetValueForIndexLooped(15))
}

func TestOrderedList_GetValueForIndexLoopedInverted(t *testing.T) {
	l := OrderedList[int]{10, 11, 12, 13, 14, 15, 16}

	require.Equal(t, 10, l.GetValueForIndexLoopedReverted(0))
	require.Equal(t, 13, l.GetValueForIndexLoopedReverted(3))
	require.Equal(t, 16, l.GetValueForIndexLoopedReverted(6))
	require.Equal(t, 15, l.GetValueForIndexLoopedReverted(7))
	require.Equal(t, 12, l.GetValueForIndexLoopedReverted(10))
	require.Equal(t, 14, l.GetValueForIndexLoopedReverted(15))
}

func TestOrderedList_GetIndexLeftOfValue(t *testing.T) {
	l := OrderedList[int]{10, 11, 12, 13, 14, 15, 16}

	require.Equal(t, 0, *l.GetIndexLeftOfValue(11))
	require.Equal(t, 6, *l.GetIndexLeftOfValue(10))
	require.Equal(t, 2, *l.GetIndexLeftOfValue(13))
	require.Equal(t, 5, *l.GetIndexLeftOfValue(16))
	require.Nil(t, l.GetIndexLeftOfValue(999))
}

func TestOrderedList_GetIndexRightOfValue(t *testing.T) {
	l := OrderedList[int]{10, 11, 12, 13, 14, 15, 16}

	require.Equal(t, 2, *l.GetIndexRightOfValue(11))
	require.Equal(t, 1, *l.GetIndexRightOfValue(10))
	require.Equal(t, 4, *l.GetIndexRightOfValue(13))
	require.Equal(t, 0, *l.GetIndexRightOfValue(16))
	require.Nil(t, l.GetIndexRightOfValue(999))
}
