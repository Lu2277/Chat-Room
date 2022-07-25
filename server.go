package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
	}
	return server
}

// Start 启动服务
func (s *Server) Start() {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("server启动成功")
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go s.Handler(conn)
	}
}

func (s *Server) Handler(conn net.Conn) {
	defer conn.Close()
	//用户上线
	user := NewUser(conn, s)
	user.OnLine()
	//广播上线消息
	s.Broadcast(user, "==Login==")
	//监听用户是否活跃的channel
	isAlive := make(chan bool, 1)
	buf := make([]byte, 4096)
	go func() {
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				defer user.conn.Close()
				user.OffLine()
				fmt.Printf("[%s]-%s退出了群聊！\n", user.Addr, user.Name)
				s.Broadcast(user, "退出了群聊!")
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println(err)
				return
			}
			//获取用户输入消息，n-1去掉回车符'\n'
			msg := string(buf[:n-1])
			//用户消息处理
			user.DoMessage(msg)
			isAlive <- true
		}
	}()
	for {
		select {
		case <-isAlive:
			//当前用户是活跃的，不做任何处理
		case <-time.After(30 * time.Second):
			//超时，将当前的用户强制下线
			user.OffLine()
			user.SendMsg("You are outed of timeout!")
			fmt.Printf("[%s]-%s超时退出了\n", user.Addr, user.Name)
			user.conn.Close()
			return
		}
	}
}

// Broadcast 广播业务，将公共消息广播给所有用户
func (s *Server) Broadcast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[%s]%s:%s\n", user.Addr, user.Name, msg)
	for _, user := range s.OnlineMap {
		user.SendMsg(sendMsg)
	}
}
