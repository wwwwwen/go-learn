package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// NewUser 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string, 0),
		conn:   conn,
		server: server,
	}

	//启动监听user channel的goroutine
	go user.ListenMessage()

	return user
}

// Online 用户上线业务
func (user *User) Online() {
	//用户上线，加入OnlineMap
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()

	//广播当前用户上线消息
	user.server.BroadCast(user, "log in")
}

// Offline 用户下线业务
func (user *User) Offline() {
	//用户下线，从OnlineMap去除
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

	//广播当前用户下线消息
	user.server.BroadCast(user, "log out")
}

// SendMsg 给当前用户对应的客户端发消息
func (user *User) SendMsg(msg string) {
	_, err := user.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("conn.Write err:", err.Error())
	}
}

// DoMessage 用户处理消息业务
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		user.server.mapLock.Lock()
		for _, u := range user.server.OnlineMap {
			onlineMsg := "[" + u.Addr + "]" + u.Name + ":在线\r\n"
			user.SendMsg(onlineMsg)
		}
		user.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式:rename|张三
		newName := msg[7:]

		user.server.mapLock.Lock()
		//判断name是否存在
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.server.mapLock.Unlock()
			user.SendMsg("当前用户名被使用\r\n")
		} else {
			delete(user.server.OnlineMap, user.Name)
			user.server.OnlineMap[newName] = user
			user.server.mapLock.Unlock()

			user.Name = newName
			user.SendMsg("您已更新用户名为:" + user.Name + "\r\n")
		}

	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式:to|张三|消息内容

		//获取对方用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.SendMsg("消息格式错误,正确格式是:to|张三|消息内容\r\n")
			return
		}

		//得到对方user对象
		remoteUser, ok := user.server.OnlineMap[remoteName]
		if !ok {
			user.SendMsg("用户名不存在\r\n")
			return
		}

		//获取消息
		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.SendMsg("内容为空,请重新发送\r\n")
			return
		}
		//发送消息
		remoteUser.SendMsg(user.Name + "对你说:" + content + "\r\n")

	} else {
		user.server.BroadCast(user, msg)
	}
}

// ListenMessage 监听当前User channel，有消息就发送给客户端
func (user *User) ListenMessage() {
	for {
		msg, isOpen := <-user.C
		if !isOpen {
			break
		}
		_, err := user.conn.Write([]byte(msg + "\r\n"))
		if err != nil {
			fmt.Println("conn.Write err:", err.Error())
		}
	}
}
