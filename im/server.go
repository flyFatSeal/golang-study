package main

import (
	"fmt"
	"net"
)

type Sever struct {
	Ip   string
	Port int
}

func (server *Sever) Handle(conn net.Conn) {
	remoteAddr := conn.RemoteAddr()
	fmt.Println("建立链接完成，请求来自", remoteAddr.String())
}

func (server *Sever) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", server.Ip, server.Port))
	if err != nil {
		fmt.Println("new.Listen error", err)
		return
	}
	defer listener.Close()

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
		Ip:   ip,
		Port: port,
	}

	return sever
}
