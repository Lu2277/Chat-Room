package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	//Ch     chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	remoteAddr := conn.RemoteAddr().String()
	user := &User{
		Name: remoteAddr,
		Addr: remoteAddr,
		//Ch:     make(chan string, 10),
		conn:   conn,
		server: server,
	}
	return user
}

// OnLine 用户上线
func (user *User) OnLine() {
	//操作map之前先上锁，安全操作
	user.server.mapLock.Lock()
	user.server.OnlineMap[user.Name] = user
	user.server.mapLock.Unlock()
}

// OffLine 用户下线
func (user *User) OffLine() {
	user.server.mapLock.Lock()
	delete(user.server.OnlineMap, user.Name)
	user.server.mapLock.Unlock()

}

// SendMsg 给user对应的客户端发送消息
func (user *User) SendMsg(msg string) {
	_, err := user.conn.Write([]byte(msg))
	if err != nil {
		return
	}
}

// DoMessage 用户消息处理
func (user *User) DoMessage(msg string) {
	//查询在线用户
	if msg == "who" {
		for _, u := range user.server.OnlineMap {
			OnlineMsg := "---[" + u.Addr + "]" + u.Name + ":" + "Online---" + "\n"
			user.SendMsg(OnlineMsg)
		}
		//用户修改名字
	} else if len(msg) > 7 && msg[:7] == "rename:" {
		newName := strings.Split(msg, ":")[1]
		_, ok := user.server.OnlineMap[newName]
		if ok {
			user.SendMsg("当前用户名已存在，请重新修改！\n")
		} else {
			user.server.mapLock.Lock()
			delete(user.server.OnlineMap, user.Name)
			user.Name = newName
			user.server.OnlineMap[user.Name] = user
			user.server.mapLock.Unlock()
			user.SendMsg("已更改用户名为：" + newName + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式 to|用户|消息
		//私聊模式
		user.PrivateChat(msg)
	} else {
		//公共消息，广播给所有用户
		user.server.Broadcast(user, msg)
	}
}

// PrivateChat 私聊模式
func (user *User) PrivateChat(msg string) {
	//获取要私聊用户名
	ToName := strings.Split(msg, "|")[1]
	if ToName == "" {
		user.SendMsg("私聊格式不对，请使用 to|用户|消息 格式重新输入\n")
		return
	}
	//判断用户是否存在
	ToUser, ok := user.server.OnlineMap[ToName]
	if !ok {
		user.SendMsg("该用户不存在，请检查！\n")
		return
	}
	//获取私聊的消息
	ToMsg := strings.Split(msg, "|")[2]
	if ToMsg == "" {
		user.SendMsg("私聊格消息不能为空！\n")
		return
	}
	ToUser.SendMsg(user.Name + "Talk To You:" + ToMsg + "\n")
}
