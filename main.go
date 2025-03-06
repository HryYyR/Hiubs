package main

import (
	"fmt"
	"net"
	"net/rpc"
	"raft/config"
	"raft/handler"
	"raft/raft"
	"time"
)

var index int
var raftPort int
var httpPort int
var address string
var configPath string

var PersistFileName string

func main() {
	ParseFlag()
	// 读取配置文件
	nodeList, err := config.InitConfiguration(configPath)
	if err != nil {
		panic("Read configFile Error :" + err.Error())
	}

	//构建节点
	//fmt.Printf("文件数据: %v\n", persist)
	Cluster := raft.Make(nodeList, index, PersistFileName)
	raft.SetRaftCluster(Cluster)

	//raft节点注册
	err = rpc.Register(raft.GetRaftCluster())
	if err != nil {
		panic(err)
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", raftPort))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	//启动httpAPI
	go handler.RegisterHandler(httpPort)

	fmt.Println("raft server listening at ", raftPort)
	rpc.Accept(listener)

}

func test(node *raft.Raft) {
	fmt.Println("----------------------------准备开始添加日志------------------------")
	time.Sleep(3 * time.Second)
	fmt.Println("添加日志,set i 1")
	node.Start("set i 1")
	time.Sleep(3 * time.Second)
	fmt.Println("添加日志,set k 2")
	node.Start("set k 2")
	time.Sleep(3 * time.Second)
	fmt.Println("添加日志,set j [3,4,5]")
	node.Start("set k 3")
	node.Start("set k 4")
	node.Start("set k 5")
	time.Sleep(3 * time.Second)
	fmt.Println("添加日志,set u 100")
	node.Start("set u 100")
	fmt.Println("-------------------------添加日志结束---------------------------")

}
