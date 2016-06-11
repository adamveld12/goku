package rpcd

import (
	"net"
	"net/rpc"

	"github.com/adamveld12/goku"
)

func init() {
	goku.RegisterService(newRPCd)
}

func newRPCd(config goku.Configuration, backend goku.Backend) goku.Service {
	return &rpcService{
		Log:     goku.NewLog("[rpc]"),
		addr:    config.RPC,
		backend: backend,
	}
}

type rpcService struct {
	goku.Log
	addr    string
	backend goku.Backend
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
