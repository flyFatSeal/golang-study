package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip            string
	Port          int
	OnlineUserMap map[string]User
	mapLock       sync.RWMutex
	Message       chan string
}

func (server *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Name + "]" + ":" + msg
	server.Message <- sendMsg
}

func (server *Server) ListenMessage() {
	for {
		msg := <-server.Message
		server.mapLock.Lock()
		for _, user := range server.OnlineUserMap {
			user.C <- msg
		}
		server.mapLock.Unlock()
	}
}

func (server *Server) Handle(conn net.Conn) {
	remoteAddr := conn.RemoteAddr()

	user := NewUser(conn, server)
	server.mapLock.Lock()
	server.OnlineUserMap[user.Addr] = *user
	server.mapLock.Unlock()

	fmt.Println("建立链接完成，请求来自", remoteAddr.String())

	// 添加 defer 确保资源清理
	defer func() {
		user.Offline()
		server.mapLock.Lock()
		delete(server.OnlineUserMap, user.Addr) // 从在线用户map中删除
		server.mapLock.Unlock()
		close(user.C)
		conn.Close()
	}()

	isLive := make(chan bool)

	user.Online()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err", err)
				return
			}

			msg := string(buf[:n-1])

			isLive <- true

			user.HandleMessage(msg)
		}
	}()

	for {
		select {
		case <-isLive:
		case <-time.After(time.Second * 10):
			user.SendMsg("超时踢出")
			return
		}
	}

}

func (server *Server) Start() {
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

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:            ip,
		Port:          port,
		OnlineUserMap: make(map[string]User),
		Message:       make(chan string),
	}

	return server
}
