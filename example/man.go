package main

import (
	"log"
	"time"

	"github.com/zh-five/golimit"
)

func main() {
	// 两种用法完全等价
	test1()
	test2()
}

// 第一种用法
func test1() {
	log.Println("开始测试1...")
	gl := golimit.NewGoLimit(2) //max_num(最大允许并发数)设置为2

	for i := 0; i < 10; i++ {
		i := i // 防止闭包陷阱. go1.22后可省略

		// 并发执行任务. 若当前并发数已达到最大, 则阻塞,直到有任务完成
		gl.Do(func() {
			time.Sleep(time.Second * 2)
			log.Println(i, "done")
		})
	}

	log.Println("循环结束")

	gl.WaitZero() //阻塞, 直到所有并发都完成
	log.Println("测试结束")

}

// 第二种用法
func test2() {
	log.Println("开始测试2...")
	gl := golimit.NewGoLimit(2) //max_num(最大允许并发数)设置为2

	for i := 0; i < 10; i++ {
		//并发计数加1.若 计数>=max_num, 则阻塞,直到 计数<max_num
		gl.Add()

		go func(g *golimit.GoLimit, i int) {
			defer gl.Done() //并发计数减1

			time.Sleep(time.Second * 2)
			log.Println(i, "done")
		}(gl, i)
	}

	log.Println("循环结束")

	gl.WaitZero() //阻塞, 直到所有并发都完成
	log.Println("测试结束")

}
