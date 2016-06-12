package app

import "github.com/adamveld12/goku"

// NewRPC creates a new RPC wrapper for a Manager
func NewRPC(b goku.Backend) *RPCWrapper {
	return &RPCWrapper{}
}

type RPCWrapper struct{}
