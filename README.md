# 一个限制go语言协程并发数的库

协程的并发数若不限制, 数量太多可能会耗尽某些资源而出现错误. 一般的做法是通过一个带缓冲区的通道来实现,缓存区的大小就是允许并发数的上限,
创建协程时读取或写入通道,退出协程是则反操作,利用通道的读或写阻塞,限制并发数的上限. 本库未采用这种方案, 
而是使用了两个sync.Mutex锁实现了并发数限制. 这样可以在运行期间调整并发上限.

### 一.快速使用
1.安装
```bash 
go get github.com/zh-five/golimit
```

2.示例
```go 
package main

import (
	"github.com/zh-five/golimit"
	"log"
	"time"
)

func main() {
	log.Println("开始测试...")
	g := golimit.NewGoLimit(2) //max_num(最大允许并发数)设置为2

	for i := 0; i < 10; i++ {
		//并发计数加1.若 计数>=max_num, 则阻塞,直到 计数<max_num
		g.Add()

		go func(g *golimit.GoLimit, i int) {
			defer g.Done() //并发计数减1

			time.Sleep(time.Second * 2)
			log.Println(i, "done")
		}(g, i)
	}


	log.Println("循环结束")

	g.WaitZero() //阻塞, 直到所有并发都完成
	log.Println("测试结束")

}


```

### 二.方法介绍

|            方法　             |                          介绍　                          |
|:-----------------------------|:--------------------------------------------------------|
|`g := golimit.NewGoLimit(200)`|创建一个对象, 设置并发上限为200　　                           |
|`g.Add()`                     |并发计数加1.若 计数>=max_num, 则阻塞,直到 计数<max_num        |
|`g.Done()`                    |并发计数减1　                                              |
|`g.WaitZero()`                |若当前并发计数为0, 则快速返回; 否则阻塞等待,直到并发计数为0　     |
|`g.SetMax(300)`               |更新最大并发计数为300, 若是调大, 可以使原阻塞的Add()快速解除阻塞　|
|`g.Count()`                   |获取当前的并发计数                                          |
|`g.Max()`                     |获取最大并发计数　                                          |



