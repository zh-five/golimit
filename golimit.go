// 协程并发数限制库
// https://github.com/zh-five/golimit

package golimit

import (
	"log"
	"sync"
)

type GoLimit struct {
	max       uint             //并发最大数量
	count     uint             //当前已有并发数
	isAddLock bool             //是否已锁定增加
	zeroChan  chan interface{} //为0时广播
	addLock   sync.Mutex       //(增加并发数的)锁
	dataLock  sync.Mutex       //(修改数据的)锁
}

func NewGoLimit(max uint) *GoLimit {
	return &GoLimit{max: max, count: 0, isAddLock: false, zeroChan: nil}
}

// 并发执行任务. 若当前并发数已达到最大, 则阻塞,直到有任务完成
func (sf *GoLimit) Do(task func()) {
	sf.Add()
	go func() {
		defer func() {
			defer sf.Done()
			if err := recover(); err != nil {
				log.Printf("golimit task panic: %v", err)
			}
		}()
		task()
	}()
}

// 并发计数加1.若 计数>=max_num, 则阻塞,直到 计数<max_num
func (sf *GoLimit) Add() {
	sf.addLock.Lock()
	sf.dataLock.Lock()

	sf.count += 1

	if sf.count < sf.max { //未超并发时解锁,后续可以继续增加
		sf.addLock.Unlock()
	} else { //已到最大并发数, 不解锁并标记. 等数量减少后解锁
		sf.isAddLock = true
	}

	sf.dataLock.Unlock()
}

// 并发计数减1
// 若计数<max_num, 可以使原阻塞的Add()快速解除阻塞
func (sf *GoLimit) Done() {
	sf.dataLock.Lock()

	sf.count -= 1

	//解锁
	if sf.isAddLock == true && sf.count < sf.max {
		sf.isAddLock = false
		sf.addLock.Unlock()
	}

	//0广播
	if sf.count == 0 && sf.zeroChan != nil {
		close(sf.zeroChan)
		sf.zeroChan = nil
	}

	sf.dataLock.Unlock()
}

// 更新最大并发计数为, 若是调大, 可以使原阻塞的Add()快速解除阻塞
func (sf *GoLimit) SetMax(n uint) {
	sf.dataLock.Lock()

	sf.max = n

	//解锁
	if sf.isAddLock == true && sf.count < sf.max {
		sf.isAddLock = false
		sf.addLock.Unlock()
	}

	//加锁
	if sf.isAddLock == false && sf.count >= sf.max {
		sf.isAddLock = true
		sf.addLock.Lock()
	}

	sf.dataLock.Unlock()
}

// 若当前并发计数为0, 则快速返回; 否则阻塞等待,直到并发计数为0
func (sf *GoLimit) WaitZero() {
	sf.dataLock.Lock()

	//无需等待
	if sf.count == 0 {
		sf.dataLock.Unlock()
		return
	}

	//无广播通道, 创建一个
	if sf.zeroChan == nil {
		sf.zeroChan = make(chan interface{})
	}

	//复制通道后解锁, 避免从nil读数据
	c := sf.zeroChan
	sf.dataLock.Unlock()

	<-c
}

// 获取并发计数
func (sf *GoLimit) Count() uint {
	return sf.count
}

// 获取最大并发计数
func (sf *GoLimit) Max() uint {
	return sf.max
}
