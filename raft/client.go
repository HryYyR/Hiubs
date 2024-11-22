package raft

import (
	"fmt"
	"log"
	"net/rpc"
)

func main() {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("Dialing:", err)
	}

	args := &Args{7, 6}
	var reply int
	err = client.Call("Arithmetic.Multiply", args, &reply)
	if err != nil {
		log.Fatal("Arithmetic error:", err)
	}
	fmt.Printf("Result: %d\n", reply)
}
