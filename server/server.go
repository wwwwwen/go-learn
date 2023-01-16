package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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

// ListenMessage 监听Message广播消息channel的goroutine，有消息就广播
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

// BroadCast 广播消息的方法
func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	server.Message <- sendMsg
}

func (server *Server) Handler(conn net.Conn) {
	//业务
	fmt.Println("连接建立成功 与", conn.RemoteAddr().String())

	user := NewUser(conn, server)
	//用户上线
	user.Online()

	//用户是否活跃
	isLive := make(chan bool)

	//接受客户端消息
	go func() {

		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err.Error())
				return
			}

			//提取消息
			msg := string(buf[0 : n-2])

			//广播消息
			user.DoMessage(msg)

			//用户是活跃的
			isLive <- true
		}
	}()

	//超时强退
	for {
		select {
		case <-isLive:
			//用户是活跃的
			//重置定时器
		case <-time.After(time.Second * 10):
			//超时
			//踢出user
			user.SendMsg("你被踢了\r\n")
			//user.Offline()这个方法会被上面的代码调用
			//销毁资源
			time.Sleep(time.Millisecond * 100)
			fmt.Println("断开连接 与", conn.RemoteAddr().String())
			close(user.C)
			err := conn.Close()
			if err != nil {
				fmt.Println("TCP connection close err:", err.Error())
			}
			return
		}
	}
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
