package goku

import (
	"net"
	"net/rpc"

	. "github.com/adamveld12/goku"
)

func init() {
	RegisterService(newRPCd)
}

func newRPCd(config Configuration, backend Backend) Service {
	return &rpcService{
		Log:     NewLog("[rpc]", config.Debug),
		addr:    config.RPC,
		backend: backend,
	}
}

type rpcService struct {
	Log
	addr    string
	backend Backend
	l       net.Listener
}

func (r *rpcService) Start() error {
	r.Trace("listening for RPC on ", r.addr)
	l, err := net.Listen("tcp", r.addr)
	if err != nil {
		return err
	}

	r.l = l
	go func(r *rpcService) {
		rpcsrv := rpc.NewServer()
		//    rpcsrv.Register()

		rpcsrv.Accept(r.l)
	}(r)

	return nil
}

func (r *rpcService) Stop() error {
	return r.l.Close()
}
