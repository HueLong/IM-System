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

	//在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

//创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}

	return server
}

//启动服务器
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen错误：", err)
		return
	}
	//close listen socket
	defer listener.Close()

	//开启协程对消息一直监听
	go this.ListenMessage()

	//accept
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept错误：", err)
			continue
		}
		//do handler
		go this.Handler(conn)
	}
}

func (this *Server) Handler(conn net.Conn) {
	user := NewUser(conn)
	//用户上线，将用户假如到OnlineMap中
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	fmt.Println("当前已上线用户数：", len(this.OnlineMap))

	//监听用户发送的消息，并广播
	go this.ReadUserMessage(conn, user)

	//广播当前用户上线消息
	this.BroadCast(user, "已上线")

	//当前Handler阻塞
	select {}

}

//广播消息
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

//监听message,将消息发送给所有用户
func (this *Server) ListenMessage() {
	for {
		msg := <- this.Message

		//将msg发送全部用户
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap{
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this * Server) ReadUserMessage(conn net.Conn, user *User) {
	buf := make([]byte, 4096)
	for  {
		n, err := conn.Read(buf)
		if n == 0 {
			this.BroadCast(user, "下线")
			fmt.Println("有用户下线啦：", user.Name)
			return
		}
		if err != nil && err != io.EOF {
			fmt.Println("Conn Read err:", err)
		}

		msg := string(buf)

		this.BroadCast(user, msg)
	}
}
