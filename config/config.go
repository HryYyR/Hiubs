package config

import (
	"encoding/json"
	"fmt"
	"os"
	"raft/raft"
)

type raftNodeConfig struct {
	NodeList []raftNode `json:"node_list"`
}

type raftNode struct {
	NodeIndex    int    `json:"node_index"`
	NodeIp       string `json:"node_ip"`
	NodeRaftPort int    `json:"node_raft_port"`
	NodeHttpPort int    `json:"node_http_port"`
	NodeTcpPort  int    `json:"node_tcp_port"`
}

type DefaultConfig struct {
	Index           int
	RaftPort        int
	HttpPort        int
	TcpPort         int
	Address         string
	ConfigPath      string
	PersistFileName string
}

var DefaultCfg *DefaultConfig = &DefaultConfig{}

// InitConfiguration 初始化配置
func InitConfiguration(cfgPath string) ([]*raft.RpcNode, error) {
	fmt.Println(cfgPath)
	if len(cfgPath) == 0 {
		return nil, fmt.Errorf("cfgPath is null")
	}
	file, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var nodeCfg raftNodeConfig
	err = json.Unmarshal(file, &nodeCfg)
	if err != nil {
		return nil, err
	}
	//fmt.Println(nodeCfg)
	var res []*raft.RpcNode
	for _, node := range nodeCfg.NodeList {
		res = append(res, &raft.RpcNode{
			Index:    node.NodeIndex,
			Address:  node.NodeIp,
			RaftPort: node.NodeRaftPort,
			HttpPort: node.NodeHttpPort,
			TcpPort:  node.NodeTcpPort,
		})
	}

	if DefaultCfg.Index == -1 {
		return nil, fmt.Errorf("node_index can be 0, use: -i assign node_index")
	} else {
		for _, node := range res {
			if node.Index == DefaultCfg.Index {
				DefaultCfg.Index = node.Index
				DefaultCfg.RaftPort = node.RaftPort
				DefaultCfg.HttpPort = node.HttpPort
				DefaultCfg.TcpPort = node.TcpPort
				DefaultCfg.Address = node.Address
				DefaultCfg.ConfigPath = cfgPath
			}
		}
	}

	return res, nil
}
