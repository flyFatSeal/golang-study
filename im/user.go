package main

import (
	"net"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func (user *User) Listener() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}

func (user *User) Online() {
	user.server.BroadCast(user, "online")
}

func (user *User) Offline() {
	user.server.BroadCast(user, "offline")
}

func (user *User) SendMsg(msg string) {
	user.conn.Write([]byte(msg + "\n"))

}

func (user *User) HandleMessage(msg string) {
	switch msg {
	case "who":
		for _, v := range user.server.OnlineUserMap {
			onlineMsg := v.Name + "  online..."
			user.SendMsg(onlineMsg)
		}
	default:
		user.server.BroadCast(user, msg)
	}
}

func NewUser(conn net.Conn, server *Server) *User {
	user := &User{
		Name:   conn.RemoteAddr().String(),
		Addr:   conn.RemoteAddr().String(),
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.Listener()
	return user
}
