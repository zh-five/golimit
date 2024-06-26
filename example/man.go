package main

import (
	"log"
	"time"

	"github.com/zh-five/golimit"
)

func main() {
	log.Println("开始测试...")
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
