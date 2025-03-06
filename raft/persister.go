package raft

import (
	"fmt"
	"os"
	"sync"
)

type Persister struct {
	mu        sync.Mutex
	raftstate []byte
	snapshot  []byte
	filename  string
}

func (ps *Persister) Save(raftstate []byte, snapshot []byte, me int) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.raftstate = clone(raftstate)
	ps.snapshot = clone(snapshot)

	err := os.WriteFile(fmt.Sprintf("./%s%d.txt", ps.filename, me), raftstate, 0644)
	if err != nil {
		fmt.Println("写入文件失败:", err)
		return
	}
	fmt.Println("日志写入成功！")

}

func clone(orig []byte) []byte {
	x := make([]byte, len(orig))
	copy(x, orig)
	return x
}
