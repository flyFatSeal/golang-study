package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Sever struct {
	Ip            string
	Port          int
	OnlineUserMap map[string]User
	mapLock       sync.RWMutex
	Message       chan string
}

func (sever *Sever) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Name + "]" + ":" + msg
	sever.Message <- sendMsg
}

func (server *Sever) ListenMessage() {
	for {
		msg := <-server.Message
		server.mapLock.Lock()
		for _, user := range server.OnlineUserMap {
			user.C <- msg
		}
		server.mapLock.Unlock()
	}
}

func (server *Sever) Handle(conn net.Conn) {
	remoteAddr := conn.RemoteAddr()

	user := NewUser(conn)
	server.mapLock.Lock()
	server.OnlineUserMap[user.Name] = *user
	server.mapLock.Unlock()

	fmt.Println("建立链接完成，请求来自", remoteAddr.String())

	server.BroadCast(user, "online alert")

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				server.BroadCast(user, "outline")
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err", err)
				return
			}

			msg := string(buf[:n])
			fmt.Println("msg::", msg)

			server.BroadCast(user, msg)
		}
	}()

}

func (server *Sever) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("new.Listen error", err)
		return
	}
	defer listener.Close()

	go server.ListenMessage()

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("listener accept error", err)
			continue
		}

		go server.Handle(conn)

	}

}

func NewSever(ip string, port int) *Sever {
	sever := &Sever{
		Ip:            ip,
		Port:          port,
		OnlineUserMap: make(map[string]User),
		Message:       make(chan string),
	}

	return sever
}
