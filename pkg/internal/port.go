package internal

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func GetFreeLocalPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	assert.True(t, ok)
	port := addr.Port

	return port
}
