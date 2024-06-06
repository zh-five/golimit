// 协程并发数限制库
// https://github.com/zh-five/golimit

package golimit

import (
	"sync"
	"sync/atomic"
)

type GoLimit struct {
	max   uint64 //并发最大数量
	count uint64 //add数

	addLock  sync.Mutex //add 等待
	waitLock sync.Mutex //zero 等待
}

func NewGoLimit(max uint64) *GoLimit {
	return &GoLimit{max: max, count: 0}
}

// 计数加1.若 计数>max_num, 则阻塞,直到 计数<max_num
func (sf *GoLimit) Add() {
	count := atomic.AddUint64(&sf.count, 1)
	max := atomic.LoadUint64(&sf.max)

	// 有协程启动，开始阻塞 waitZero()
	if count == 1 {
		sf.waitLock.Lock()
	}
	if count >= max {
		sf.addLock.Lock()
	}
}

// 并发计数减1
// 若计数<max_num, 可以使原阻塞的Add()快速解除阻塞
func (sf *GoLimit) Done() {
	count := atomic.AddUint64(&sf.count, ^uint64(0))
	max := atomic.LoadUint64(&sf.max)

	// 所有并发协程结束，让 waitZero() 解除阻塞
	if count == 0 {
		sf.waitLock.Unlock()
	}

	//
	if count+1 >= max {
		sf.addLock.Unlock()
	}
}

// 更新最大并发计数为, 若是调大, 可以使原阻塞的Add()快速解除阻塞
func (sf *GoLimit) SetMax(max uint64) {
	oldMax := atomic.LoadUint64(&sf.max)
	atomic.StoreUint64(&sf.max, max)

	count := atomic.AddUint64(&sf.count, ^uint64(0))

	if oldMax == max {
		// ok
	} else if oldMax < max {
		if count >= oldMax { // 有解锁处理
			// 最大程度解除已阻塞的add
			end := count
			if count > max {
				end = max
			}
			for i := oldMax; i < end; i++ {
				sf.addLock.Unlock()
			}

			if count < max { // 解除预锁定
				sf.addLock.Unlock()
			}
		}
	} else if oldMax > max { // 加锁处理
		if count < oldMax && count >= max { // 原max未锁定，新max需锁定
			sf.addLock.Lock()
		}
	}
}

// 若当前并发计数为0, 则快速返回; 否则阻塞等待,直到并发计数为0
func (sf *GoLimit) WaitZero() {
	sf.waitLock.Lock()
	defer sf.waitLock.Unlock()
}

// 获取并发计数
func (sf *GoLimit) Count() uint64 {
	n := atomic.LoadUint64(&sf.count)
	max := atomic.LoadUint64(&sf.max)
	if n > max {
		return max
	}

	return n
}

// 获取最大并发计数
func (sf *GoLimit) Max() uint64 {
	return atomic.LoadUint64(&sf.max)
}
