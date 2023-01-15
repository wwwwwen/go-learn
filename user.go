package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

// 创建一个用户的API
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string, 0),
		conn: conn,
	}

	//启动监听user channel的goroutine
	go user.ListenMessage()

	return user
}

// 监听当前User channel，有消息就发送给客户端
func (user *User) ListenMessage() {
	for {
		msg, isOpen := <-user.C
		if !isOpen {
			break
		}
		user.conn.Write([]byte(msg + "\r\n"))
	}
}
