package main

import (
	"fmt"
	"net"
	"net/rpc"
	"raft/config"
	"raft/handler"
	"raft/raft"
)

func main() {
	ParseFlag()
	// 读取配置文件
	nodeList, err := config.InitConfiguration(config.DefaultCfg.ConfigPath)
	cfg := config.DefaultCfg
	if err != nil {
		panic(fmt.Sprintf("Read configFile Error:%s or not assign node_index use: -i assign node_index", err.Error()))
	}

	//构建节点
	//fmt.Printf("文件数据: %v\n", persist)
	Cluster := raft.Make(nodeList, cfg.Index, cfg.PersistFileName)
	raft.SetRaftCluster(Cluster)

	//raft节点注册
	err = rpc.Register(raft.GetRaftCluster())
	if err != nil {
		panic(err)
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.RaftPort))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	//启动httpAPI
	//启动TCP服务器
	//启动raft节点通信RPC服务
	go handler.RegisterHandler(cfg.HttpPort)
	go handler.StartTCPServer(cfg.TcpPort)
	fmt.Println("raft server listening at ", cfg.RaftPort)
	rpc.Accept(listener)
}
