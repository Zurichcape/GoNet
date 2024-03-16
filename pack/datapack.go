package pack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"gonet/config"
	"gonet/interfaces"
)

var defaultHeaderLen uint32 = 8

// DataPack 拆包、封包的具体模块
type DataPack struct {
}

func NewDataPack() *DataPack {
	return &DataPack{}
}

// GetHeadLen 获取包的头部长度
func (d *DataPack) GetHeadLen() uint32 {
	//DataLen(4字节)+IDLen(4字节)
	return 8
}

// Pack 封包方法
func (d *DataPack) Pack(msg interfaces.IMessage) ([]byte, error) {
	//存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	//将dataLen写入dataBuff中
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgLen()); err != nil {
		return nil, err
	}
	//将MsgID写入dataBuff中
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}
	//将data数据写入dataBUff中
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}
	return dataBuff.Bytes(), nil
}

// UnPack 拆包方法, 首先读取head信息，之后再根据head信息里的data长度再进行一次读
func (d *DataPack) UnPack(binaryData []byte) (interfaces.IMessage, error) {
	//创建一个读二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)
	msg := &Message{}
	//首先解压head信息，得到dataLen和MsgID
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.ID); err != nil {
		return nil, err
	}
	//判断是否已经超出了允许的MaxPackageSize
	if config.GlobalServerConfig.MaxPacketSize > 0 && msg.DataLen > config.GlobalServerConfig.MaxPacketSize {
		return nil, errors.New("msg beyond the limitation")
	}
	//这里只需要把head的数据拆包出来就可以了，然后再通过head的长度，再从conn读取一次数据
	return msg, nil
}
