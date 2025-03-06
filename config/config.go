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
}

// InitConfiguration 初始化配置
func InitConfiguration(cfgPath string) ([]*raft.RpcNode, error) {
	fmt.Println(cfgPath)
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
		})
	}

	return res, nil
}
