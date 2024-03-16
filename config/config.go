// Package utils 提供相关工具类函数
// 包括:
//
//	全局配置
//	配置文件加载
//
// 相关配置文件定义及加载方式
package config

import (
	"github.com/go-ini/ini"
	"os"

	interfaces "gonet/interfaces"
)

/*
存储一切全局参数，供其他模块使用
一些参数也可以通过server.json来配置
*/
type GlobalObj struct {
	/*
		server
	*/
	TCPServer interfaces.IServer // 当前全局Server对象
	IPVersion string             // IP版本
	Host      string             // 当前服务器主机IP
	TCPPort   int                // 当前服务器主机监听端口号
	Name      string             // 当前服务器名称

	Version          string // 当前服务版本号
	MaxPacketSize    uint32 // 都需数据包的最大值
	MaxConn          int    // 当前服务器主机允许的最大链接个数
	WorkerPoolSize   uint   // 业务工作Worker池的数量
	MaxWorkerTaskLen uint32 // 业务工作Worker对应负责的任务队列最大任务存储数量
	MaxMsgChanLen    uint32 // SendBuffMsg发送消息的缓冲最大长度

	/*
		config file path
	*/
	ConfFilePath string
	ConfigFile   *ini.File

	/*
		fluentd
	*/
	FluentdHost      string
	FluentdPort      int
	FluentdDebugMode bool
}

/*
定义一个全局的对象
*/
var GlobalServerConfig *GlobalObj

// PathExists 判断一个文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// Reload 读取用户的配置文件
func (g *GlobalObj) Reload() {
	if confFileExists, _ := PathExists(g.ConfFilePath); !confFileExists {
		panic("config file " + g.ConfFilePath + " is not exist!!")
		return
	}

	file, err := ini.Load(g.ConfFilePath)
	if err != nil {
		panic("load config file " + g.ConfFilePath + " failed:" + err.Error())
		return
	}
	// 缓存配置给其他parser读取
	g.ConfigFile = file

	parseServer(g, file)
	parseFluentd(g, file)
}

/*
提供init方法，默认加载
*/
func init() {
	pwd, err := os.Getwd()
	if err != nil {
		pwd = "."
	}
	// 初始化GlobalObject变量，设置配置文件路径
	GlobalServerConfig = &GlobalObj{
		ConfFilePath: pwd + "/conf/server.ini",
	}
	// 从配置文件中加载一些用户配置的参数
	GlobalServerConfig.Reload()
}

// 读取服务器配置信息
func parseServer(config *GlobalObj, file *ini.File) {
	section := file.Section("Server")
	config.Name = section.Key("Name").MustString("hoxi-server")
	config.Host = section.Key("Host").MustString("0.0.0.0")
	config.TCPPort = section.Key("TCPPort").MustInt(8999)
	config.IPVersion = section.Key("IPVersion").MustString("tcp4")
	config.Version = section.Key("Version").MustString("V1")
	config.MaxPacketSize = uint32(section.Key("MaxPacketSize").MustUint(4096))
	config.MaxConn = section.Key("MaxConn").MustInt(12000)
	config.WorkerPoolSize = section.Key("WorkerPoolSize").MustUint(10)
	config.MaxWorkerTaskLen = uint32(section.Key("MaxWorkerTaskLen").MustUint(1024))
	config.MaxMsgChanLen = uint32(section.Key("MaxMsgChanLen").MustUint(1024))
}

// 读取Fluentd配置
func parseFluentd(config *GlobalObj, file *ini.File) {
	section := file.Section("Fluentd")
	config.FluentdHost = section.Key("Host").MustString("127.0.0.1")
	config.FluentdPort = section.Key("Port").MustInt(24224)
	config.FluentdDebugMode = section.Key("DebugMode").MustBool(true)
}
