package raft

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
)

var (
	FUNC_GET = "get"
	FUNC_SET = "set"
	FUNC_DEL = "del"
	FUNC_ALL = "all"
)

type StateMachine struct {
	Data *sync.Map
}

var CommandList = []string{
	"get",
	"set",
	"del",
	"all",
}

var UnPersistCommand = []string{
	"get",
	"all",
}

func (this *StateMachine) get(key string) (string, bool) {
	value, ok := this.Data.Load(key)
	return value.(string), ok
}

func (this *StateMachine) set(key, value string) {
	this.Data.Store(key, value)
}

func (this *StateMachine) all() map[string]string {
	var res = make(map[string]string)
	this.Data.Range(func(key, value any) bool {
		res[key.(string)] = value.(string)
		return true
	})
	return res
}

func (this *StateMachine) HandleCommand(command string) (string, bool) {
	//是否为合法命令
	commandSplit := strings.Split(command, " ")
	if len(commandSplit) < 1 {
		fmt.Println("非法指令(长度错误)", command)
		return "", false
	}

	contains := slices.Contains(CommandList, commandSplit[0])
	if !contains {
		fmt.Println("非法指令(非法命令)", command)
		return "", false
	}

	switch commandSplit[0] {
	case FUNC_GET:
		if len(commandSplit) != 2 {
			return "", false
		}
		value, ok := this.get(commandSplit[1])
		return value, ok
	case FUNC_SET:
		if len(commandSplit) != 3 {
			return "", false
		}
		this.set(commandSplit[1], commandSplit[2])
		return "", true
	case FUNC_DEL:
		if len(commandSplit) != 2 {
			return "", false
		}
		this.Data.Delete(commandSplit[1])
		return "", true
	case FUNC_ALL:
		all := this.all()
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
			rf.StateMachine.HandleCommand(applyMsg.Command.(string))
		}
	}

}
