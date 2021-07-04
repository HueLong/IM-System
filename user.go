package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C: make(chan string),
		conn: conn,
	}
	//开启协程，对用户channel开启监听
	go user.ListenMessage()

	return user
}

//监听当前User channel的方法，一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for  {
		msg := <- this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
