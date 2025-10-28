package raft

import (
	"fmt"
	"net/rpc"
)

type RpcNode struct {
	Client   *rpc.Client
	Address  string
	RaftPort int
	HttpPort int
	TcpPort  int
	Index    int
}

func (r *RpcNode) GetClient() (*rpc.Client, error) {
	client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", r.Address, r.RaftPort))
	if err != nil {
		return nil, err
	}

	r.Client = client
	return r.Client, nil
}

func (r *RpcNode) Ticker(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}
