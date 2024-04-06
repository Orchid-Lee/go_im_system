package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	//在线用户列表
	OnlineUsrMap map[string]*User
	mapLock      sync.RWMutex
	//消息广播
	Message chan string
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:           ip,
		Port:         port,
		OnlineUsrMap: make(map[string]*User),
		Message:      make(chan string),
	}
	return server
}

// 监听Message广播消息channel的goroutine, 一旦有消息发送给全部在线用户
func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message
		//将msg发送给全部在线用户
		server.mapLock.Lock()
		for _, cli := range server.OnlineUsrMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

// 消息广播方法
func (server *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendMsg
}

// 业务处理方法
func (server *Server) Handler(conn net.Conn) {
	//...当前链接的业务
	fmt.Println("登录成功~")
	//用户上线, 将用户加到OnlineMap
	user := NewUser(conn)
	server.mapLock.Lock()
	server.OnlineUsrMap[user.Name] = user
	server.mapLock.Unlock()

	//广播当前消息
	server.Broadcast(user, "已上线")
	//携程无父子关系, 父携程中开启新携程, 父携程退出, 不影响子携程

	//接受用户客户端消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				server.Broadcast(user, "下线~")
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read Err:", err)
			}

			msg := string(buf[:n-1])
			//将得到的消息广播
			server.Broadcast(user, msg)
		}

	}()
}

// 启动服务器的接口
func (server *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("net.Listen() err:", err)
	}
	//close socket
	defer listener.Close()

	//启动监听Messaged的goroutine
	go server.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept() err:", err)
			continue
		}
		//do handler
		go server.Handler(conn)
	}
}
