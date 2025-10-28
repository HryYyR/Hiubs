package main

import (
	"flag"
	"fmt"
	"raft/config"
)

func ParseFlag() {
	flag.IntVar(&config.DefaultCfg.Index, "i", -1, "索引")
	flag.StringVar(&config.DefaultCfg.ConfigPath, "config", "./config/node_config.json", "配置文件")
	flag.IntVar(&config.DefaultCfg.RaftPort, "raft_port", 8080, "http端口号")
	flag.IntVar(&config.DefaultCfg.HttpPort, "http_port", 1234, "raft端口号")
	flag.IntVar(&config.DefaultCfg.TcpPort, "tcp_port", 8998, "tcp端口号")
	flag.StringVar(&config.DefaultCfg.Address, "a", "127.0.0.1", "ip地址")
	flag.StringVar(&config.DefaultCfg.PersistFileName, "pname", "persist", "持久化文件名称前缀")
	flag.Parse()

	fmt.Println(*config.DefaultCfg)
}
