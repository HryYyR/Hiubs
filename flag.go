package main

import (
	"flag"
	"fmt"
)

func ParseFlag() {
	flag.IntVar(&index, "i", 1, "索引")
	flag.IntVar(&raftPort, "raft_port", 8080, "http端口号")
	flag.IntVar(&httpPort, "http_port", 1234, "raft端口号")
	flag.StringVar(&address, "a", "127.0.0.1", "ip地址")
	flag.StringVar(&configPath, "config", "./config/node_config.json", "配置文件")
	flag.StringVar(&PersistFileName, "pname", "persist", "持久化文件名称前缀")
	flag.Parse()

	fmt.Println(index, raftPort, httpPort, address, configPath)

}
