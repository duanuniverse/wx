/*
服务器端：
	1. 接受来自客户端的连接请求并建立连接
	2. 所以的客户端连接会放进连接池，用于广播消息

客户端：
	连接服务器
	向服务器发送消息
	接收服务器的广播消息

主要事项：
	某一个客户端断开连接后需要从连接池摘除，并不再接受广播消息
	某个客户端断开连接后不能影响服务端或别的客户端连接
*/

package main

import (
	"fmt"
	"log"
	"net"
)

//向所有的群友发广播
func BroadMessage(conns *map[string]net.Conn, messageChan chan string) {
	for{
		//不断的从通道里读取消息
		msg := <- messageChan
		fmt.Println(msg)

		//向所有的群友发送消息
		for key, conn := range *conns{
			fmt.Println("connection is connected from", key)
			_, err := conn.Write([]byte(msg))
			if err != nil{
				log.Printf("broad messge to %s failed: %v\n", key, err)
				delete(*conns, key)
			}
		}
	}
}

//处理客户端发到服务端的消息，将其扔到通道内
func Handler(conn net.Conn, conns *map[string]net.Conn, message chan string) {
	fmt.Println("connect from client", conn.RemoteAddr().String())

	buf := make([]byte, 1024)
	for {
		lenght, err := conn.Read(buf)
		if err != nil{
			log.Printf("read client message failed: %v\n", err)
			delete(*conns, conn.RemoteAddr().String())
			conn.Close()
			break
		}
		recvStr := string(buf[0:lenght])
		message <- recvStr
	}
}

func Start(port string)  {
	host := ":" + port

	tcpAddr, err := net.ResolveTCPAddr("tcp4", host)
	if err != nil{
		log.Printf("resolve tcp addr failed: %v\n", err)
		return
	}
	//监听
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil{
		log.Printf("listen tcp port failed: %v\n", err)
		return
	}
	//建立连接池，用于广播消息
	conns := make(map[string]net.Conn)
	//消息通道
	messageChan := make(chan string, 10)
	//广播消息
	go BroadMessage(&conns, messageChan)
	//启动
	for {
		fmt.Printf("listening port %s ...\n", port)
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("接收失败：%v\n", err)
			continue
		}
		//把每个客户端连接放进连接池
		conns[conn.RemoteAddr().String()] = conn
		fmt.Println(conns)

		//处理消息
		go Handler(conn, &conns, messageChan)
	}
}

func main() {
	port := "9090"
	Start(port)
}


