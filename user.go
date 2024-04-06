package main

import (
	"net"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

// 创建用户API
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{Name: userAddr, Addr: userAddr, C: make(chan string), conn: conn}
	//启动监听当前user channel的goroutine
	go user.ListenMessage()
	return user
}

// 监听当前user channel的方法, 有消息则发送给客户端
func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\r\n"))
		user.conn.Write([]byte(""))
	}
}
