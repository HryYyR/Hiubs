package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"raft/raft"
	"time"
)

func main() {
	var port int
	var index int
	var address string
	flag.IntVar(&port, "p", 1234, "端口号")
	flag.IntVar(&index, "i", 1, "索引")
	flag.StringVar(&address, "a", "127.0.0.1", "ip地址")
	flag.Parse()
	fmt.Println(port, index, address)

	node1 := &raft.RpcNode{Client: nil, Address: "127.0.0.1", Port: 1233}
	node2 := &raft.RpcNode{Client: nil, Address: "127.0.0.1", Port: 1234}
	node3 := &raft.RpcNode{Client: nil, Address: "127.0.0.1", Port: 1235}

	persist, err := os.ReadFile(fmt.Sprintf("./persis%d.txt", index))
	if err != nil {
		fmt.Println("读取文件失败:", err)
	}
	fmt.Printf("文件数据: %v\n", persist)
	raftnode := raft.Make([]*raft.RpcNode{node1, node2, node3}, index, persist)

	err = rpc.Register(raftnode)
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("raft server listening at ", port)

	go runhttp(raftnode, index)

	rpc.Accept(listener)

}

func runhttp(node *raft.Raft, index int) {
	http.HandleFunc("/commit", func(w http.ResponseWriter, r *http.Request) {
		cmd := r.URL.Query().Get("cmd")
		fmt.Println(cmd)
		logindex, term, success := node.Start(cmd)
		fmt.Printf("logindex: %d, term: %d, success: %v\n", logindex, term, success)
		return
	})

	port := index + 8080
	fmt.Println("http server listening at ", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func test(node *raft.Raft) {
	fmt.Println("----------------------------准备开始添加日志------------------------")
	time.Sleep(3 * time.Second)
	fmt.Println("添加日志,set i 1")
	node.Start("set i 1")
	time.Sleep(3 * time.Second)
	fmt.Println("添加日志,set k 2")
	node.Start("set k 2")
	time.Sleep(3 * time.Second)
	fmt.Println("添加日志,set j [3,4,5]")
	node.Start("set k 3")
	node.Start("set k 4")
	node.Start("set k 5")
	time.Sleep(3 * time.Second)
	fmt.Println("添加日志,set u 100")
	node.Start("set u 100")
	fmt.Println("-------------------------添加日志结束---------------------------")

}
