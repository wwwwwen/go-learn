package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//广播channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Message广播消息channel的goroutine，有消息就广播
func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message
		//广播msg
		server.mapLock.Lock()
		for _, cli := range server.OnlineMap {
			cli.C <- msg
		}
		server.mapLock.Unlock()
	}
}

// 广播消息的方法
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	//业务
	fmt.Println("连接建立成功 与", conn.RemoteAddr().String())

	user := NewUser(conn)
	//用户上线，加入OnlineMap
	server.mapLock.Lock()
	server.OnlineMap[user.Name] = user
	server.mapLock.Unlock()

	//广播当前用户上线消息
	server.BroadCast(user, "log in")

	//handler阻塞
	//select {}
}

func (server *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Printf("net.Listen err: %v\n", err)
	}
	//close listen socket
	defer listener.Close()

	//启动监听Message的goroutine
	go server.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("listener accept err: %v\n", err)
			continue
		}
		//do handler
		go server.Handler(conn)
	}

}
