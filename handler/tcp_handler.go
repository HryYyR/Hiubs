package handler

import (
	"bufio"
	"fmt"
	"net"
)

// StartTCPServer 启动 TCP 服务器
func StartTCPServer(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Failed to start TCP server: %v\n", err)
		return
	}
	defer listener.Close()
	fmt.Printf("TCP server listening on port %d\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		go handleConnection(conn)
	}
}

// handleConnection 处理客户端连接
func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		var message string
		scanner := bufio.NewScanner(reader)
		if scanner.Scan() {
			message = scanner.Text()
		}

		response, ret := HandleCommand(message)
		if ret != COMMAND_OK {
			conn.Write([]byte("INVALID COMMAND\n"))
		}
		conn.Write([]byte(response + "\n"))
	}
}
