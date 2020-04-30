package main

import (
	"github.com/zh-five/golimit"
	"log"
	"time"
)

func main() {
	log.Println("开始测试...")
	g := golimit.NewGoLimit(2)

	for i := 0; i < 10; i++ {
		g.Add()
		go func(g *golimit.GoLimit, i int) {
			defer g.Done()

			time.Sleep(time.Second * 2)
			log.Println(i, "done")
		}(g, i)
	}
	log.Println("循环结束")

	g.WaitZero()
	log.Println("测试结束")

}
