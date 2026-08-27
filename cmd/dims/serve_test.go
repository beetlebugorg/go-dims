package main

import (
	"testing"
	"time"

	"github.com/beetlebugorg/go-dims/internal/core"
	"github.com/stretchr/testify/require"
)

func TestNewServerAppliesTimeouts(t *testing.T) {
	config := *core.ReadConfig()
	config.ReadHeaderTimeout = 1000
	config.ReadTimeout = 2000
	config.WriteTimeout = 3000
	config.IdleTimeout = 4000
	config.MaxHeaderBytes = 4096
	config.BindAddress = ":9999"

	server := newServer(&config)

	require.Equal(t, ":9999", server.Addr)
	require.Equal(t, 1*time.Second, server.ReadHeaderTimeout)
	require.Equal(t, 2*time.Second, server.ReadTimeout)
	require.Equal(t, 3*time.Second, server.WriteTimeout)
	require.Equal(t, 4*time.Second, server.IdleTimeout)
	require.Equal(t, 4096, server.MaxHeaderBytes)
	require.NotNil(t, server.Handler)
}

// The defaults must be timeouts, not the zero values the http package reads
// as "wait forever".
func TestDefaultServerTimeoutsAreSet(t *testing.T) {
	server := newServer(core.ReadConfig())

	require.NotZero(t, server.ReadHeaderTimeout)
	require.NotZero(t, server.ReadTimeout)
	require.NotZero(t, server.WriteTimeout)
	require.NotZero(t, server.IdleTimeout)
	require.NotZero(t, server.MaxHeaderBytes)
}

func TestMilliseconds(t *testing.T) {
	require.Equal(t, 5*time.Second, milliseconds(5000))
	require.Equal(t, time.Duration(0), milliseconds(0), "zero must stay zero")
}
