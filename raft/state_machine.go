package raft

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
)

var (
	FUNC_GET        = "get"
	FUNC_SET        = "set"
	FUNC_DEL        = "del"
	FUNC_ALL        = "all"
	FUNC_REGISTER   = "register"
	FUNC_HEART_BEAT = "heartbeat"
)

type StateMachine struct {
	Data *sync.Map
}

var CommandList = []string{
	"get",
	"set",
	"del",
	"all",
	"register",
	"heartbeat",
}

var UnPersistCommand = []string{
	"get",
	"all",
	"heartbeat",
}

func (this *StateMachine) Get(key string) (string, bool) {
	value, ok := this.Data.Load(key)
	return value.(string), ok
}

func (this *StateMachine) Set(key, value string) {
	this.Data.Store(key, value)
}

func (this *StateMachine) All() map[string]string {
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
		value, ok := this.Get(commandSplit[1])
		return value, ok
	case FUNC_SET:
		if len(commandSplit) != 3 {
			return "", false
		}
		this.Set(commandSplit[1], commandSplit[2])
		return "", true
	case FUNC_DEL:
		if len(commandSplit) != 2 {
			return "", false
		}
		this.Data.Delete(commandSplit[1])
		return "", true
	case FUNC_ALL:
		all := this.All()
		allBytes, _ := json.Marshal(all)
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
