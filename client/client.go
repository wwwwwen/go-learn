package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	opt        int //client的状态
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建客户端
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		opt:        -1,
	}

	//连接服务器
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	return client
}

// DealResponse 处理server回应的消息，直接显示到标准输出
func (client *Client) DealResponse() {
	_, err := io.Copy(os.Stdout, client.conn)
	if err != nil {
		fmt.Println("func DealResponse err:", err.Error())
		return
	}
}

func (client *Client) menu() bool {
	var opt int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("4.查询所有用户")
	fmt.Println("0.退出")
	_, _ = fmt.Scanln(&opt)

	if opt >= 0 && opt <= 4 {
		client.opt = opt
		return true
	} else {
		fmt.Println("请输入合法数字")
		return false
	}
}

func (client *Client) PublicChat() bool {
	//提示输入消息
	var chatMsg string

	fmt.Println("请输入聊天内容,exit退出")
	_, _ = fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//不为空就发给服务器
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\r\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err.Error())
				client.opt = 0
				return false
			}
		}
		chatMsg = ""
		// FIXME
		// 如果用户在这里被踢出 必须至少输入两次字符才能退出程序
		_, _ = fmt.Scanln(&chatMsg)
	}
	return true
}

func (client *Client) UpdateName() bool {
	fmt.Println("请输入用户名")
	_, _ = fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\r\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err", err.Error())
		client.opt = 0
		return false
	}
	return true
}

func (client *Client) QueryAllUsers() bool {
	_, err := client.conn.Write([]byte("who\r\n"))
	if err != nil {
		fmt.Println("conn.Write err", err.Error())
		client.opt = 0
		return false
	}
	return true
}

func (client *Client) Run() {
	for client.opt != 0 {
		for client.menu() != true {
		}

		//根据不同模式处理不同业务
		switch client.opt {
		case 1:
			//公聊模式
			fmt.Println("公聊模式选择")
			client.PublicChat()
		case 2:
			//私聊模式
			fmt.Println("私聊模式选择")
		case 3:
			//更新用户名
			fmt.Println("更改用户名模式选择")
			client.UpdateName()
		case 4:
			fmt.Println("查询所有用户")
			client.QueryAllUsers()
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "h", "127.0.0.1", "设置服务器ip地址,默认为127.0.0.1")
	flag.IntVar(&serverPort, "p", 8888, "设置服务器端口,默认为8888")
}

func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("连接服务器失败...")
	}

	fmt.Println("连接服务器成功...")

	//单独开启goroutine处理server消息
	go client.DealResponse()

	//启动客户端业务
	client.Run()
}
