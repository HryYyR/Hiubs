package HiubsClient

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// HiubsClient 用于与 Raft 集群的 TCP 服务交互
type HiubsClient struct {
	conn       net.Conn
	addr       string        // 服务器地址（如 "localhost:8081"）
	timeout    time.Duration // 超时时间
	maxRetries int           // 最大重试次数
}

// NewHiubsClient 创建一个新的 HiubsClient 实例
func NewHiubsClient(addr string) *HiubsClient {
	return &HiubsClient{
		addr:       addr,
		timeout:    5 * time.Second, // 默认超时时间
		maxRetries: 3,               // 默认最大重试次数
	}
}

// Connect 建立与 TCP 服务器的连接
func (c *HiubsClient) Connect() error {
	var conn net.Conn
	var err error

	for i := 0; i < c.maxRetries; i++ {
		conn, err = net.DialTimeout("tcp", c.addr, c.timeout)
		if err == nil {
			c.conn = conn
			return nil
		}

		// 重试前等待一段时间（指数退避）
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return fmt.Errorf("failed to connect to server after %d retries: %v", c.maxRetries, err)
}

// Disconnect 关闭连接
func (c *HiubsClient) Disconnect() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendCommand 发送命令并返回响应
func (c *HiubsClient) SendCommand(command string) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("connection is not established")
	}

	var response string
	var err error

	for i := 0; i < c.maxRetries; i++ {
		// 设置读写超时
		err = c.conn.SetDeadline(time.Now().Add(c.timeout))
		if err != nil {
			return "", fmt.Errorf("failed to set deadline: %v", err)
		}

		// 发送命令
		_, err = fmt.Fprintf(c.conn, "%s\n", command)
		if err != nil {
			continue
		}

		// 读取响应
		reader := bufio.NewReader(c.conn)
		response, err = reader.ReadString('\n')
		if err == nil {
			return strings.TrimSpace(response), nil
		}

		// 重试前等待一段时间（指数退避）
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return "", fmt.Errorf("failed to send command after %d retries: %v", c.maxRetries, err)
}

// Add 添加键值对
func (c *HiubsClient) Add(key, value string) (string, error) {
	return c.SendCommand(fmt.Sprintf("add %s %s", key, value))
}

// Delete 删除键值对
func (c *HiubsClient) Delete(key string) (string, error) {
	return c.SendCommand(fmt.Sprintf("delete %s", key))
}

// Update 更新键值对
func (c *HiubsClient) Update(key, value string) (string, error) {
	return c.SendCommand(fmt.Sprintf("update %s %s", key, value))
}

// Get 获取键值对
func (c *HiubsClient) Get(key string) (string, error) {
	return c.SendCommand(fmt.Sprintf("get %s", key))
}

// List 列出所有键值对
func (c *HiubsClient) List() (string, error) {
	return c.SendCommand("list")
}
