package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server
}

func (server *Server)Handler(conn net.Conn){
	//业务
	fmt.Println("链接建立成功")
}

func (server *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Printf("net.Listen err: %v\n", err)
	}
	//close listen socket
	defer listener.Close()
	for {
		//accept
		conn, err := listener.Accept()
		if err!=nil{
			fmt.Printf("listener accept err: %v\n", err)
			continue
		}
		//do handler
		go server.Handler(conn)
	}

	
}
