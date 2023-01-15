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

	//接受客户端消息
	//这个goroutine好像有点多余
	go func() {
		//close connection
		defer func(conn net.Conn) {
			fmt.Println("断开连接 与", conn.RemoteAddr().String())
			err := conn.Close()
			if err != nil {
				fmt.Println("TCP connection close err:", err.Error())
			}
		}(conn)
		char := make([]byte, 1)
		for {
			buf := make([]byte, 0, 4096)
			//windows telnet的bug，一次只能发一个字符
			//而且ctrl+]然后send sth. 不会发最后的回车
			for {
				num, err := conn.Read(char)
				if num == 0 {
					server.BroadCast(user, "log out")
					return
				}
				if err != nil && err != io.EOF {
					fmt.Println("Conn Read err:", err.Error())
					return
				}
				//但是回车竟然是发两个字符 \r\num
				if char[0] == '\r' || char[0] == '\n' {
					break
				} else {
					buf = append(buf, char[0])
				}
			}
			if len(buf) == 0 || char[0] == '\n' {
				continue
			}
			//提取消息
			msg := string(buf[0:])

			//广播消息
			server.BroadCast(user, msg)

		}
	}()

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
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("listen socket close err:", err.Error())
		}
	}(listener)

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
