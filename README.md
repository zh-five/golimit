# 一个限制go语言协程并发数的库

协程的并发数若不限制, 数量太多可能会耗尽某些资源而出现错误. 一般的做法是通过一个带缓冲区的通道来实现,缓存区的大小就是允许并发数的上限,
创建协程时读取或写入通道,退出协程是则反操作,利用通道的读或写阻塞,限制并发数的上限. 本库未采用这种方案, 
而是使用了两个sync.Mutex锁实现了并发数限制. 这样设计可获得以下一些优势：
1. 增加并发限制数量不会增加内存消耗
2. 运行期间可调整并发限制数量
3. 可阻塞等待所有协程运行结束
4. 可实时获取当前并发协程数量

### 一.快速使用
1.安装
```bash 
go get github.com/zh-five/golimit
```

2.示例
`example/man.go`

```go 
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



```

### 二.方法介绍

|            方法　             |                          介绍　                          |
|:-----------------------------|:--------------------------------------------------------|
|`g := golimit.NewGoLimit(200)`|创建一个对象, 设置并发上限为200　　                           |
|`g.Add()`                     |并发计数加1.若 计数>=max_num, 则阻塞,直到 计数<max_num        |
|`g.Done()`                    |并发计数减1　                                              |
|`g.Do(task func())`           |并发执行任务，自动调用`g.Add()`和`g.Done()`　                |
|`g.WaitZero()`                |若当前并发计数为0, 则快速返回; 否则阻塞等待,直到并发计数为0　     |
|`g.SetMax(300)`               |更新最大并发计数为300, 若是调大, 可以使原阻塞的Add()快速解除阻塞　|
|`g.Count()`                   |获取当前的并发计数                                          |
|`g.Max()`                     |获取最大并发计数　                                          |


### 三.使用场景介绍
#### 场景1
给可能出现大量并发的异步任务增加并发限制, 防止服务器资源耗尽.
例如: 一个web服务接口, 每次请求都会启动一个协程处理任务, 若并发数过多, 可能会导致服务器资源消耗过多, 所以需要限制并发数.
通常做法是启动和维护一个channel和若干任务处理协程，改造成本有些高，使用golimit库则可以很方便的实现。
```go

// 服务处理方法
func (sf *Service) ApiHandel(/* args .. */) {
	// 其它处理 。。。

	// 启动异步任务
	go task()

	// 返回数据 。。。
}


```
改造后
```go

func NewService() *Service {
	return &Service{
		gl: golimit.NewGoLimit(200), // 设置最大并发数为200
	}
}

// 服务处理方法
func (sf *Service) ApiHandel(/* args .. */) {
	// 其它处理 。。。

	// 异步调用gl.Do()，是为了接口快速返回
	go func() {
		// 有并发上限的启动异步任务
	    gl.Do(func() {
			task()
		})
	}()

	// 返回数据 。。。
}

```

#### 场景2
快速改造串行任务为并发任务，并限制并发数。 
例如批量下载文件，原来为串行下载，现在希望加快速度，但又不想并发数过多导致服务器资源耗尽。
```go

func downloadFile(url string) {
	// 下载文件
}

// 串行下载
func main() {
	for _, url := range urls {
		downloadFile(url)
	}
	fmt.Println("下载完成")
}

// 改造为并发下载
func main() {
	gl := golimit.NewGoLimit(5) // 设置最大并发数为5
	for _, url := range urls {
		url := url
		gl.Do(func() {
			downloadFile(url)
		})
	}
	gl.WaitZero() // 阻塞, 直到所有并发都完成
	fmt.Println("下载完成")
}
```