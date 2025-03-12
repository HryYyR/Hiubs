package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"raft/raft"
	"slices"
	"strings"
)

var (
	COMMAND_INVALID uint32 = 0
	COMMAND_OK      uint32 = 1
)

type CommandResponse struct {
	Data string
	Ret  uint32
}

func commitService(w http.ResponseWriter, r *http.Request) {
	var res CommandResponse
	cmd := r.URL.Query().Get("cmd")
	cmd, valid := CheckCommandValid(cmd)
	if !valid {
		res.Ret = COMMAND_INVALID
		resByte, _ := json.Marshal(res)
		w.Write(resByte)
		return
	}

	cmdsplit := strings.Split(cmd, " ")

	if !slices.Contains(raft.CommandList, cmdsplit[0]) {
		res.Ret = COMMAND_INVALID
		resByte, _ := json.Marshal(res)
		w.Write(resByte)
		return
	}

	if slices.Contains(raft.UnPersistCommand, cmdsplit[0]) {
		data, success := raft.GetRaftCluster().StateMachine.HandleCommand(cmd)
		if !success {
			res.Ret = COMMAND_INVALID
			resByte, _ := json.Marshal(res)
			w.Write(resByte)
			return
		}
		res.Data = data
	} else {
		_, _, success := raft.GetRaftCluster().Start(cmd)
		if !success {
			res.Ret = COMMAND_INVALID
			resByte, _ := json.Marshal(res)
			w.Write(resByte)
			return
		}
	}

	res.Ret = COMMAND_OK
	resByte, err := json.Marshal(res)
	if err != nil {
		fmt.Println("marshal error", err)
	}
	w.Write(resByte)
}

func allStateService(w http.ResponseWriter, r *http.Request) {
	var res = make(map[string]string)
	raft.GetRaftCluster().StateMachine.Data.Range(func(key, value any) bool {
		res[key.(string)] = value.(string)
		return true
	})
	resbyte, _ := json.Marshal(res)
	w.Write(resbyte)
}
