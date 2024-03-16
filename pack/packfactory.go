package pack

import (
	"gonet/interfaces"
	"sync"
)

var packOnce sync.Once

type packFactory struct{}

var FactoryInstance *packFactory

/*
Factory 生成不同的解包方法
*/
func init() {
	packOnce.Do(func() {
		FactoryInstance = new(packFactory)
	})
}

// NewPack 创建一个具体的拆包解包对象
// 工厂方法的设计模式
func (f *packFactory) NewPack(kind string) interfaces.IDataPack {
	var dataPack interfaces.IDataPack
	switch kind {
	case interfaces.GoNetDataPack:
		dataPack = NewDataPack()
	default:
		dataPack = NewDataPack()
	}
	return dataPack
}
