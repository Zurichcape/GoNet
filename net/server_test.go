package net

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func ClientTest() {
	fmt.Println("Client Test ... start")
	//阻塞三秒后发起测试请求，给服务端留出开始服务的时间
	time.Sleep(3 * time.Second)
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Println("client start err: ", err)
		return
	}
	for {
		_, err = conn.Write([]byte("hello zinx_v1.0"))
		if err != nil {
			fmt.Println("write error: ", err)
			return
		}
		buf := make([]byte, 512)
		cnt, err := conn.Read(buf)
		if err != nil {
			fmt.Println("recv buf error: ", err)
			return
		}
		fmt.Printf("server call back: %s, cnt = %d\n", buf, cnt)
		time.Sleep(1 * time.Second)
	}
}

func TestServer(t *testing.T) {
	/*
		服务端测试
	*/
	//创建一个server句柄
	s := NewServer()
	s.Start()
	/*
		客户端测试
	*/
	go ClientTest()
}
