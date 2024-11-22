package raft

import (
	"net"
	"net/rpc"
)

type Arithmetic int

type Args struct {
	A, B int
}

func (t *Arithmetic) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func server() {
	arith := new(Arithmetic)
	rpc.Register(arith)
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	rpc.Accept(listener)
}
