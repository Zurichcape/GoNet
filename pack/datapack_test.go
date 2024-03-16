package pack

import (
	"fmt"
	"io"
	"net"
	"testing"
)

func TestDataPack(t *testing.T) {
	/*
		模拟的服务器
	*/
	listener, err := net.Listen("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Println("server listen err: ", err)
		return
	}
	//创建服务器goroutine，负责从客户端goroutine读取粘包的数据，然后进行解析
	go func() {
		/*处理客户端的请求
		 *---->拆包的过程<------
		 *定义一个拆包的对象dp
		 */
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("server accept err: ", err)
				return
			}
			go func(conn net.Conn) {
				dp := NewDataPack()
				for {
					//1.第一次从conn读，把包的head读出来
					headData := make([]byte, dp.GetHeadLen())
					_, err := io.ReadFull(conn, headData)
					if err != nil {
						fmt.Println("server read head err: ", err)
						break
					}
					msgHead, err := dp.UnPack(headData)
					if err != nil {
						fmt.Println("server unpack err: ", err)
						return
					}
					//如果成功解析出数据
					if msgHead.GetMsgLen() > 0 {
						msg := msgHead.(*Message)
						msg.Data = make([]byte, msg.GetMsgLen())

						//2.第二次从conn读，把包的data读出来
						_, err := io.ReadFull(conn, msg.Data)
						if err != nil {
							fmt.Println("server unpack data err: ", err)
							return
						}
						fmt.Println("==> Recv Msg: ID=", msg.ID, "len=", msg.DataLen, "data=", string(msg.Data))
					}

				}
			}(conn)
		}
	}()

	/*
	 *模拟客户端
	 */

	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Println("client dial err: ", err)
		return
	}

	//封装一个消息
	dp := NewDataPack()
	msg1 := &Message{
		ID:      0,
		DataLen: 5,
		Data:    []byte{'h', 'e', 'l', 'l', 'o'},
	}
	sendData1, err := dp.Pack(msg1)
	if err != nil {
		fmt.Println("client pack msg1 err: ", err)
		return
	}
	msg2 := &Message{
		ID:      1,
		DataLen: 7,
		Data:    []byte{'w', 'o', 'r', 'l', 'd', '!', '!'},
	}
	sendData2, err := dp.Pack(msg2)
	if err != nil {
		fmt.Println("client pack msg2 err: ", err)
		return
	}
	sendData1 = append(sendData1, sendData2...)
	//一次性发送给服务端
	_, err = conn.Write(sendData1)
	if err != nil {
		fmt.Println("client send msg err: ", err)
		return
	}
	//客户端则色
	select {}
}
