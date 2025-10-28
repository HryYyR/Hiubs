package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"raft/raft"
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

	response, ret := HandleCommand(cmd)
	res.Ret = ret
	res.Data = response
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
