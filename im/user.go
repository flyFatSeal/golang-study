package main

import (
	"net"
	"strings"
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
	command, args := user.extractCommand(msg)

	switch command {
	case "who":
		for _, v := range user.server.OnlineUserMap {
			onlineMsg := v.Name + "  online..."
			user.SendMsg(onlineMsg)
		}
	case "rename":
		name := args[0]
		for _, v := range user.server.OnlineUserMap {
			if v.Name == name {
				user.SendMsg("当前名称已被占用，请更换!")
				return
			}
		}
		user.server.mapLock.Lock()
		userInfo := user.server.OnlineUserMap[user.Addr]
		userInfo.Name = name
		user.server.OnlineUserMap[user.Addr] = userInfo
		user.Name = name
		user.server.mapLock.Unlock()
		user.SendMsg("更换名字成功，您当前用户名为:" + name)
	default:
		user.server.BroadCast(user, msg)
	}
}

func (user *User) extractCommand(msg string) (string, []string) {
	if strings.HasPrefix(msg, "rename|") {
		newName := strings.SplitN(msg, "|", 2)[1]
		return "rename", []string{newName}
	}
	return msg, []string{}
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
