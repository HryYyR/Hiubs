package raft

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"net/rpc"
	"os"
	"raft/labgob"
	"sync"
	"sync/atomic"
	"time"
)

var raftCluster *Raft

const (
	Follower = iota
	Candidate
	Leader
)

var HeartBeatTimeOut = 200
var CommitCheckTimeInterval = 1000

// 在 3D 部分中，你将需要通过 applyCh 发送其他类型的消息（例如快照），
// 但对于这些用途，请将 CommandValid 设置为 false。
// ApplyMsg 结构体用于将提交的日志条目或快照传递给服务层。

type ApplyMsg struct {
	CommandValid bool        // 表示 Command 是否为新提交的日志条目
	Command      interface{} // 新提交的日志命令
	CommandIndex int         // 新提交日志的索引

	// 适用于 3D 部分（用于快照的字段）：
	SnapshotValid bool   // 表示 Snapshot 是否为有效快照
	Snapshot      []byte // 快照数据
	SnapshotTerm  int    // 快照的任期
	SnapshotIndex int    // 快照的日志索引
}

type Entry struct {
	Term int
	Cmd  interface{}
}

type Raft struct {
	StateMachine sync.Map
	mu           sync.Mutex // 锁，保护该节点状态的并发访问
	peers        []*RpcNode // 所有节点的 RPC 终端
	persister    *Persister // 用于保存该节点的持久状态
	me           int        // 节点在 peers[] 中的索引
	dead         int32      // 通过 Kill() 设置，标记节点是否已被终止

	// 用于存放 Raft 节点的状态数据（在 3A、3B、3C 中实现）。
	// 详见论文图 2 中 Raft 服务器需要维护的状态描述。
	// state a Raft server must maintain.

	currentTerm int
	votedFor    int
	log         []Entry

	commitIndex int           // 将要提交的日志的最高索引
	lastApplied int           // 已经被应用到状态机的日志的最高索引
	nextIndex   []int         // 复制到某一个follower时, log开始的索引
	matchIndex  []int         // 已经被复制到follower的日志的最高索引
	applyCh     chan ApplyMsg // 用于在应用到状态机时传递消息

	// 以下不是Figure 2中的field
	timeStamp time.Time // 记录收到消息的时间(心跳或append)
	role      int

	muVote sync.Mutex // 保护投票数据
	//voteCount int
}

type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int // candidate’s term
	CandidateId  int // candidate requesting vote
	LastLogIndex int // index of candidate’s last log entry (§5.4)
	LastLogTerm  int // term of candidate’s last log entry (§5.4)
}

// RequestVoteReply RPC 回复结构。
type RequestVoteReply struct {
	Term        int  // currentTerm, for candidate to update itself
	VoteGranted bool // true means candidate received vote
}

type AppendEntriesArgs struct {
	// Your data here (2A, 2B).
	Term         int     // leader’s term
	LeaderId     int     // so follower can redirect clients
	PrevLogIndex int     // index of log entry immediately preceding new ones
	PrevLogTerm  int     // term of prevLogIndex entry
	Entries      []Entry // log entries to store (empty for heartbeat; may send more than one for efficiency)
	LeaderCommit int     // leader’s commitIndex
}

func GetRaftCluster() *Raft {
	return raftCluster
}
func SetRaftCluster(cluster *Raft) *Raft {
	raftCluster = cluster
	return raftCluster
}

func (rf *Raft) persist() {
	fmt.Printf("server %v 开始持久化, 最后一个持久化的log为: %v:%v", rf.me, len(rf.log)-1, rf.log[len(rf.log)-1].Cmd)
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(rf.votedFor)
	e.Encode(rf.currentTerm)
	e.Encode(rf.log)
	raftstate := w.Bytes()
	rf.persister.Save(raftstate, nil, rf.me)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil {
		return
	}
	if len(data) < 1 { // bootstrap without any state?
		return
	}
	if len(data) == 0 {
		return
	}
	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)

	var votedFor int
	var currentTerm int
	var log []Entry
	if d.Decode(&votedFor) != nil ||
		d.Decode(&currentTerm) != nil ||
		d.Decode(&log) != nil {
		panic("读取持久化数据失败！\n")
	} else {
		rf.votedFor = votedFor
		rf.currentTerm = currentTerm
		rf.log = log
		//rf.persist()
		for _, entry := range rf.log {
			if entry.Cmd == nil {
				continue
			}
			rf.HandleCommand(entry.Cmd.(string))
		}
	}
}

// Start 新增一条日志
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	// 如果不是leader返回false
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.role != Leader {
		return -1, -1, false
	}

	newEntry := &Entry{Term: rf.currentTerm, Cmd: command}
	rf.log = append(rf.log, *newEntry)
	return len(rf.log) - 1, rf.currentTerm, true
}

type AppendEntriesReply struct {
	// Your data here (2A).
	Term    int  // currentTerm, for leader to update itself
	Success bool // true if follower contained entry matching prevLogIndex and prevLogTerm
}

// RequestVote 处理投票
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	// Your code here (2A, 2B).
	rf.mu.Lock()

	if args.Term < rf.currentTerm {
		// 旧的term
		// 1. Reply false if term < currentTerm (§5.1)
		reply.Term = rf.currentTerm
		rf.mu.Unlock()
		reply.VoteGranted = false
		fmt.Printf("server %v 拒绝向 server %v投票: 旧的term: %v,\n\targs= %+v\n", rf.me, args.CandidateId, args.Term, args)
		return nil
	}

	// 代码到这里时, args.Term >= rf.currentTerm

	//如果有节点大于自己的任期，更新自己的任期，并退出选举
	if args.Term > rf.currentTerm {
		// 已经是新一轮的term, 之前的投票记录作废
		rf.votedFor = -1
		rf.currentTerm = args.Term // 易错点, 需要将currentTerm提升到最新的term
		rf.role = Follower
		rf.persist()
	}

	// at least as up-to-date as receiver’s log, grant vote (§5.2, §5.4)
	if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
		// 首先确保是没投过票的
		if args.LastLogTerm > rf.log[len(rf.log)-1].Term ||
			(args.LastLogTerm == rf.log[len(rf.log)-1].Term && args.LastLogIndex >= len(rf.log)-1) {
			// 2. If votedFor is null or candidateId, and candidate’s log is least as up-to-date as receiver’s log, grant vote (§5.2, §5.4)
			rf.currentTerm = args.Term
			reply.Term = rf.currentTerm
			rf.votedFor = args.CandidateId
			rf.role = Follower
			rf.timeStamp = time.Now()
			rf.persist()

			rf.mu.Unlock()
			reply.VoteGranted = true
			fmt.Printf("server %v 同意向 server %v投票 args= %+v\n", rf.me, args.CandidateId, args)
			return nil
		}
	} else {
		fmt.Printf("server %v 拒绝向 server %v投票: 已投票 args= %+v\n", rf.me, args.CandidateId, rf.votedFor)
	}

	reply.Term = rf.currentTerm
	rf.mu.Unlock()
	reply.VoteGranted = false
	return nil
}

func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// 如果需要，在此处添加其他代码。
}

// 检查节点是否已被 Kill() 调用。
func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func Make(peers []*RpcNode, index int, PersistFileName string) *Raft {
	rf := &Raft{}

	//读取持久化文件
	persist, err := os.ReadFile(fmt.Sprintf("./%s%d.txt", PersistFileName, index))
	if err != nil {
		fmt.Println("读取文件失败:", err)
		create, err := os.Create(fmt.Sprintf("./%s%d.txt", PersistFileName, index))
		if err != nil {
			panic(err)
		}
		create.Close()
	}
	persister := &Persister{
		filename: PersistFileName,
	}
	rf.persister = persister

	rf.StateMachine = sync.Map{}
	rf.peers = peers
	rf.me = index
	rf.role = Follower
	rf.applyCh = make(chan ApplyMsg, 1024)

	rf.readPersist(persist)

	rf.nextIndex = make([]int, len(peers))
	rf.matchIndex = make([]int, len(peers))

	if len(rf.log) == 0 {
		rf.log = make([]Entry, 0)
		rf.log = append(rf.log, Entry{
			Term: 0,
			Cmd:  nil,
		})
	} else {
		for i := 0; i < len(rf.nextIndex); i++ {
			rf.nextIndex[i] = len(rf.log)
		}
	}

	go rf.ticker()
	go rf.CommitChecker()
	go rf.ApplyState()

	return rf
}

func (rf *Raft) CommitChecker() {
	// 检查是否有新的commit
	for !rf.killed() {
		rf.mu.Lock()
		for rf.commitIndex > rf.lastApplied {
			rf.lastApplied += 1
			msg := &ApplyMsg{
				CommandValid: true,
				Command:      rf.log[rf.lastApplied].Cmd,
				CommandIndex: rf.lastApplied,
			}
			fmt.Printf("server %v 准备将命令 %v(索引为 %v ) 应用到状态机\n", rf.me, msg.Command, msg.CommandIndex)
			rf.applyCh <- *msg
		}
		rf.mu.Unlock()
		time.Sleep(time.Duration(CommitCheckTimeInterval) * time.Millisecond)
	}
}

// SendHeartBeats 发送心跳
func (rf *Raft) SendHeartBeats() {
	fmt.Printf("server %v 开始发送心跳\n", rf.me)

	for !rf.killed() {
		rf.mu.Lock()
		// if the server is dead or is not the leader, just return
		if rf.role != Leader {
			rf.mu.Unlock()
			// 不是leader则终止心跳的发送
			fmt.Println("不是leader则终止心跳的发送")
			return
		}

		for i := 0; i < len(rf.peers); i++ {
			if i == rf.me {
				continue
			}

			args := &AppendEntriesArgs{
				Term:         rf.currentTerm,
				LeaderId:     rf.me,
				PrevLogIndex: rf.nextIndex[i] - 1,            //该节点日志，索引的位置
				PrevLogTerm:  rf.log[rf.nextIndex[i]-1].Term, //该节点最后一个索引日志的任期
				LeaderCommit: rf.commitIndex,                 //leader提交到的日志的索引（已提交的最新的日志的索引）
			}

			//如果该节点的最后一条日志的索引和leader的不同，说明该子节点有未同步的日志
			if len(rf.log)-1 >= rf.nextIndex[i] {
				// 如果有新的log需要发送, 则就是一个真正的AppendEntries而不是心跳
				args.Entries = rf.log[rf.nextIndex[i]:]
				fmt.Printf("leader %v 开始向 server %v 广播新的AppendEntries,日志内容：%v\n", rf.me, i, args.Entries)
			} else {
				// 如果没有新的log发送, 就发送一个长度为0的切片, 表示心跳
				args.Entries = make([]Entry, 0)
				//fmt.Printf("leader %v 开始向 server %v 广播新的心跳, args = %+v \n", rf.me, i, args)
			}

			go rf.handleAppendEntries(i, args)
		}
		rf.mu.Unlock()

		time.Sleep(time.Duration(HeartBeatTimeOut) * time.Millisecond)
	}

	//defer func() {
	//	if r := recover(); r != nil {
	//		fmt.Println("panic: ", r)
	//	}
	//}()
}

// 发送 (心跳/日志) rpc消息
func (rf *Raft) handleAppendEntries(serverTo int, args *AppendEntriesArgs) {
	//fmt.Println("发送心跳rpc ", serverTo, args)
	reply := &AppendEntriesReply{}
	sendArgs := *args // 复制一份args结构体, (可能有未知的错误)
	ok := rf.sendAppendEntries(serverTo, &sendArgs, reply)
	//fmt.Println(ok)

	if !ok {
		return
	}

	rf.mu.Lock()
	defer func() {
		// DPrintf("server %v handleAppendEntries 释放锁mu", rf.me)
		rf.mu.Unlock()
	}()

	if sendArgs.Term != rf.currentTerm {
		// 函数调用间隙值变了
		return
	}

	//如果日志消息接收成功
	if reply.Success {
		rf.nextIndex[serverTo] = rf.matchIndex[serverTo] + 1            //更新该节点的最新日志索引位置
		rf.matchIndex[serverTo] = args.PrevLogIndex + len(args.Entries) // 更新该节点的日志的最高索引

		// 需要判断是否可以commit
		N := len(rf.log) - 1

		//通过检查每个节点的最新日志索引，判断日志是否可以提交
		for N > rf.commitIndex {
			count := 1 // 1表示包括了leader自己，统计大于该索引的节点的个数
			for i := 0; i < len(rf.peers); i++ {
				if i == rf.me {
					continue
				}
				if rf.matchIndex[i] >= N && rf.log[N].Term == rf.currentTerm {
					count += 1
				}
			}
			if count > len(rf.peers)/2 {
				// 如果至少一半的follower回复了成功, 更新commitIndex
				break
			}
			N -= 1
		}

		rf.commitIndex = N

		return
	}

	if reply.Term > rf.currentTerm {
		// 回复了比自己新的term, 表示自己已经不是leader了
		fmt.Printf("server %v 旧的leader收到了心跳函数中更新的term: %v, 转化为Follower\n", rf.me, reply.Term)

		rf.currentTerm = reply.Term
		rf.role = Follower
		rf.votedFor = -1
		rf.timeStamp = time.Now()
		rf.mu.Unlock()
		rf.persist()
		return
	}

	//如果任期和角色没有问题，但是success返回失败，说明传过去的日志与子节点的日志的索引不匹配，尝试逐渐自减重试，直到成功
	if reply.Term == rf.currentTerm && rf.role == Leader {
		fmt.Println("如果任期和角色没有问题，但是success返回失败，说明传过去的日志与子节点的日志的索引不匹配，尝试逐渐自减重试，直到成功")
		// term仍然相同, 且自己还是leader, 表名对应的follower在prevLogIndex位置没有与prevLogTerm匹配的项
		// 将nextIndex自减再重试
		rf.nextIndex[serverTo] -= 1
		rf.mu.Unlock()
		return
	}
}

// AppendEntries 接收心跳和日志
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	// Your code here (2A, 2B).
	// 新leader发送的第一个消息
	rf.mu.Lock()

	//如果leader的任期比自己小，就拒绝接收心跳，等待超时选举
	if args.Term < rf.currentTerm {
		// 这是来自旧的leader的消息
		// 1. Reply false if term < currentTerm (§5.1)
		fmt.Println("leader的任期比自己小，就拒绝接收心跳，等待超时选举")
		reply.Term = rf.currentTerm
		rf.mu.Unlock()
		reply.Success = false
		return nil
	}

	// 代码执行到这里就是 args.Term >= rf.currentTerm 的情况

	// 不是旧 leader的话需要记录访问时间
	rf.timeStamp = time.Now()

	if args.Term > rf.currentTerm {
		// 新leader的第一个消息
		rf.currentTerm = args.Term // 更新iterm
		rf.votedFor = -1           // 易错点: 更新投票记录为未投票
		rf.role = Follower
		rf.persist()
	}

	if args.Entries == nil {
		// 心跳函数
		//fmt.Printf("server %v 接收到 leader &%v 的心跳\n", rf.me, args.LeaderId)
	} else {
		fmt.Printf("server %v 收到 leader %v 的的AppendEntries: %+v \n", rf.me, args.LeaderId, args)
	}

	//日志部分

	// 校验PrevLogIndex和PrevLogTerm不合法
	if args.Entries != nil &&
		(args.PrevLogIndex >= len(rf.log) || rf.log[args.PrevLogIndex].Term != args.PrevLogTerm) {

		reply.Term = rf.currentTerm
		rf.mu.Unlock()
		reply.Success = false
		fmt.Printf("server %v 检查到心跳中参数不合法:\n\t args.PrevLogIndex=%v, args.PrevLogTerm=%v, \n\tlen(self.log)=%v, self最后一个位置term为:%v\n", rf.me, args.PrevLogIndex, args.PrevLogTerm, len(rf.log), rf.log[len(rf.log)-1].Term)
		return nil
	}

	//如果是日志心跳，并且自己的日志长度比leader的长，并且最后一条日志的任期和leader的任期不同
	//追加日志
	for idx, log := range args.Entries {
		ridx := args.PrevLogIndex + 1 + idx
		if ridx < len(rf.log) && rf.log[ridx].Term != log.Term {
			// 某位置发生了冲突, 覆盖这个位置开始的所有内容
			rf.log = rf.log[:ridx]
			rf.log = append(rf.log, args.Entries[idx:]...)
			rf.persist()
			break
		} else if ridx == len(rf.log) {
			// 没有发生冲突但长度更长了, 直接拼接
			rf.log = append(rf.log, args.Entries[idx:]...)
			rf.persist()
			break
		}
	}

	if len(args.Entries) != 0 {
		fmt.Printf("server %v 成功进行apeend,%v \n", rf.me, rf.log)
	}

	reply.Success = true
	reply.Term = rf.currentTerm

	if args.LeaderCommit > rf.commitIndex {
		// 5.If leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
		rf.commitIndex = int(math.Min(float64(args.LeaderCommit), float64(len(rf.log)-1)))
		fmt.Println("更新commitIndex 为", rf.commitIndex)
	}
	rf.mu.Unlock()
	return nil
}

// 检查心跳
func (rf *Raft) ticker() {
	for !rf.killed() {
		//检查上次接收心跳到当前时间是否超过了400-500毫秒
		rdTimeOut := rand.Intn(100) + 400
		time.Sleep(time.Duration(rand.Intn(50)+200) * time.Millisecond) //每200-250毫秒检查一次是否超时

		rf.mu.Lock()
		if rf.role != Leader && time.Since(rf.timeStamp) > time.Duration(rdTimeOut)*time.Millisecond {
			// 选举超时，触发选举
			rf.mu.Unlock()
			fmt.Println("超时，开始选举", rf.me)
			go rf.Elect()
		} else {
			rf.mu.Unlock()
		}
	}
}

// Elect 开始选举
func (rf *Raft) Elect() {
	var muVote sync.Mutex // 临时的投票锁
	muVote.Lock()
	rf.mu.Lock()

	rf.currentTerm += 1 // 自增term
	rf.role = Candidate // 成为候选人
	rf.votedFor = rf.me // 给自己投票
	//rf.voteCount = 1          // 自己有一票
	voteCount := 1            // 自己有一票
	rf.timeStamp = time.Now() // 自己给自己投票也算一种消息
	rf.persist()

	args := &RequestVoteArgs{
		Term:         rf.currentTerm,
		CandidateId:  rf.me,
		LastLogIndex: len(rf.log) - 1,
		LastLogTerm:  rf.log[len(rf.log)-1].Term,
		//LastLogTerm: 0,
	}
	rf.mu.Unlock()
	muVote.Unlock()

	for i := 0; i < len(rf.peers); i++ {
		if i == rf.me {
			continue
		}
		go rf.collectVote(i, args, &muVote, &voteCount)
	}
}

// 收集选票
func (rf *Raft) collectVote(serverTo int, args *RequestVoteArgs, muVote *sync.Mutex, voteCount *int) {
	voteAnswer := rf.GetVoteAnswer(serverTo, args)
	//fmt.Println(serverTo, voteAnswer)
	if !voteAnswer {
		return
	}
	muVote.Lock()
	if *voteCount > len(rf.peers)/2 {
		muVote.Unlock()
		return
	}

	*voteCount += 1
	fmt.Println("获得票数", voteCount, len(rf.peers)/2)
	if *voteCount > len(rf.peers)/2 {
		rf.mu.Lock()
		if rf.role == Follower {
			// 有另外一个投票的协程收到了更新的term而更改了自身状态为Follower
			rf.mu.Unlock()
			muVote.Unlock()
			return
		}
		fmt.Println(rf.me, "当选leader")
		rf.role = Leader
		// 需要重新初始化nextIndex和matchIndex
		for i := 0; i < len(rf.nextIndex); i++ {
			rf.nextIndex[i] = len(rf.log)
			rf.matchIndex[i] = 0
		}
		rf.mu.Unlock()
		go rf.SendHeartBeats()
	}

	muVote.Unlock()
}

func (rf *Raft) GetVoteAnswer(server int, args *RequestVoteArgs) bool {
	sendArgs := *args
	reply := RequestVoteReply{}
	ok := rf.sendRequestVote(server, &sendArgs, &reply)
	if !ok {
		return false
	}

	rf.mu.Lock()
	defer rf.mu.Unlock()

	if sendArgs.Term != rf.currentTerm {
		// 易错点: 函数调用的间隙被修改了
		return false
	}

	if reply.Term > rf.currentTerm {
		// 已经是过时的term了
		rf.currentTerm = reply.Term
		rf.votedFor = -1
		rf.role = Follower
		rf.persist()
	}
	return reply.VoteGranted
}

// 发送心跳/追加日志的rpc
func (rf *Raft) sendAppendEntries(index int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	client, err := rf.peers[index].GetClient()
	if err != nil {
		//fmt.Println("获取", index, "客户端失败：", err)
		return false
	}
	defer func(client *rpc.Client) {
		_ = client.Close()
	}(client)
	err = client.Call("Raft.AppendEntries", args, reply)
	if err != nil {
		return false
	}
	return true
}

// 发送选举投票的rpc
func (rf *Raft) sendRequestVote(index int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	client, err := rf.peers[index].GetClient()
	if err != nil {
		fmt.Println("获取", index, "客户端失败：", err)
		return false
	}
	defer func(client *rpc.Client) {
		_ = client.Close()
	}(client)
	err = client.Call("Raft.RequestVote", args, reply)
	if err != nil {
		fmt.Println("调用", index, "rpc失败：", err)
		return false
	}
	return true
}
