package raft

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

var (
	FUNC_GET = "get"
	FUNC_SET = "set"
	FUNC_DEL = "del"
	FUNC_ALL = "all"
)

var commondList = []string{
	"get",
	"set",
	"del",
	"all",
}

func (rf *Raft) get(key string) (string, bool) {
	value, ok := rf.StateMachine.Load(key)
	return value.(string), ok
}

func (rf *Raft) set(key, value string) {
	rf.StateMachine.Store(key, value)
}

func (rf *Raft) all() map[string]string {
	var res = make(map[string]string)
	rf.StateMachine.Range(func(key, value any) bool {
		res[key.(string)] = value.(string)
		return true
	})
	return res
}

func (rf *Raft) HandleCommand(command string) (string, bool) {
	//是否为合法命令
	commandSplit := strings.Split(command, " ")
	if len(commandSplit) < 1 {
		fmt.Println("非法指令(长度错误)", command)
		return "", false
	}

	contains := slices.Contains(commondList, commandSplit[0])
	if !contains {
		fmt.Println("非法指令(非法命令)", command)
		return "", false
	}

	switch commandSplit[0] {
	case FUNC_GET:
		if len(commandSplit) != 2 {
			return "", false
		}
		value, ok := rf.get(commandSplit[1])
		return value, ok
	case FUNC_SET:
		if len(commandSplit) != 3 {
			return "", false
		}
		rf.set(commandSplit[1], commandSplit[2])
		return "", true
	case FUNC_DEL:
		if len(commandSplit) != 2 {
			return "", false
		}
		rf.StateMachine.Delete(commandSplit[1])
		return "", true
	case FUNC_ALL:
		all := rf.all()
		allBytes, _ := json.Marshal(all)
		fmt.Println(string(allBytes))
		return string(allBytes), true

	default:
		return "", false
	}
}

func (rf *Raft) ApplyState() {

	for !rf.killed() {
		select {
		case applyMsg := <-rf.applyCh:
			rf.HandleCommand(applyMsg.Command.(string))
		}
	}

}
