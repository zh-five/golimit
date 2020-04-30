package golimit

import "sync"

type GoLimit struct {
	max       uint       //并发最大数量
	num       uint       //当前已有并发数
	isAddLock bool       //是否已锁定增加
	addLock   sync.Mutex //(增加并发数的)锁
	dataLock  sync.Mutex //(修改数据的)锁
}

func NewGoLimit(max uint) *GoLimit {
	return &GoLimit{max: max, num: 0, isAddLock: false}
}

//开始一个新协程
//若超过数量限制,则阻塞等待.否则计数加1并不阻塞
func (g *GoLimit) Add() {
	g.addLock.Lock()
	g.dataLock.Lock()

	g.num += 1

	if g.num < g.max { //未超并发时解锁,后续可以继续增加
		g.addLock.Unlock()
	} else { //已到最大并发数, 不解锁并标记. 等数量减少后解锁
		g.isAddLock = true
	}

	g.dataLock.Unlock()
}

//一个协程完成退出时调用, 计数减1
//若有阻塞等等,且并发数低于上限,解锁
func (g *GoLimit) Done() {
	g.dataLock.Lock()

	g.num -= 1

	//解锁
	if g.isAddLock == true && g.num < g.max {
		g.isAddLock = false
		g.addLock.Unlock()
	}

	g.dataLock.Unlock()
}

func (g *GoLimit) SetMax(n uint) {
	g.dataLock.Lock()

	g.max = n

	//解锁
	if g.isAddLock == true && g.num < g.max {
		g.isAddLock = false
		g.addLock.Unlock()
	}

	//加锁
	if g.isAddLock == false && g.num >= g.max {
		g.isAddLock = true
		g.addLock.Lock()
	}

	g.dataLock.Unlock()
}
