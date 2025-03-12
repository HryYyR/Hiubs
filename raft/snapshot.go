package raft

import (
	"encoding/json"
	"fmt"
	"os"
)

type SnapShot struct {
	Index int
	Term  int
	Data  map[string]string
}

func (rf *Raft) ReadSnapShot(PersistFileName string, index int) bool {
	//读取快照文件
	data, err := os.ReadFile(fmt.Sprintf("./%s%d_snapshot.txt", PersistFileName, index))
	if err != nil {
		fmt.Println("读取文件失败 开始创建:", err)
		create, err := os.Create(fmt.Sprintf("./%s%d_snapshot.txt", PersistFileName, index))
		if err != nil {
			panic(err)
		}
		create.Close()
	}
	if len(data) == 0 {
		return false
	}
	snapShotStruct := SnapShot{}
	if err := json.Unmarshal(data, &snapShotStruct); err != nil {
		fmt.Println("反序列化快照失败:", err)
		return false
	}
	rf.SnapShot = snapShotStruct
	return true
}

func (rf *Raft) SaveSnapShot() {
	rf.SnapShot.Data = rf.StateMachine.all()
	rf.SnapShot.Term = rf.log[rf.lastApplied].Term
	rf.SnapShot.Index = rf.lastApplied
	bytes, err := json.Marshal(rf.SnapShot)
	if err != nil {
		fmt.Printf("序列化快照失败 %v\n", err)
	}
	err = os.WriteFile(fmt.Sprintf("./%s%d_snapshot.txt", rf.persister.filename, rf.me), bytes, 0644)
	if err != nil {
		fmt.Printf("写入快照失败 %v\n", err)
	}
	fmt.Printf("写入快照成功 索引位置:%d 任期:%d \n", rf.SnapShot.Index, rf.SnapShot.Term)
	return
}

type SnapShotArgs struct {
	Term     int // leader’s term
	LeaderId int // so follower can redirect clients
}

type SnapShotReply struct {
	Term        int  // currentTerm, for candidate to update itself
	VoteGranted bool // true means candidate received vote
}

func (rf *Raft) handleSnapShot(serverTo int, args *SnapShotArgs) {
}
