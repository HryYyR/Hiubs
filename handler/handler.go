package handler

import (
	"raft/raft"
	"slices"
	"strings"
)

func CheckCommandValid(command string) (string, bool) {
	result := strings.ReplaceAll(command, "  ", " ")
	result = strings.ReplaceAll(command, "\n", "")
	// 连续替换多余的空格
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}
	result = strings.ToLower(result)

	return result, true
}
func HandleCommand(command string) (string, uint32) {
	cmd, valid := CheckCommandValid(command)
	if !valid {
		return "", COMMAND_INVALID
	}

	cmdSplit := strings.Split(cmd, " ")

	if !slices.Contains(raft.CommandList, cmdSplit[0]) {
		return "", COMMAND_INVALID
	}

	if slices.Contains(raft.UnPersistCommand, cmdSplit[0]) {
		data, success := raft.GetRaftCluster().StateMachine.HandleCommand(cmd)
		if !success {
			return "", COMMAND_INVALID
		}
		return data, COMMAND_OK
	} else {
		_, _, success := raft.GetRaftCluster().Start(cmd)
		if !success {
			return "", COMMAND_INVALID
		}
		return "", COMMAND_OK
	}
}
