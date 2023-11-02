package queue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const (
	stop = iota
	run
)

// Pool 协程池子
type Pool struct {
	//   当前goroutine数量
	countWorker uint
	//最大goroutine数量
	maxWorker uint
	//	taskQueue  任务队列
	taskQueue chan func(v ...any)
	//waitQueue  等待队列
	waitQueue *Wait
	//	执行队列
	workQueue chan func(v ...any)
	//锁
	look sync.Mutex
	//	status 状态
	status int
	//执行一次
	once sync.Once
	//id
	idle bool
	//	日志
	*log.Logger
}

/*
先启动一个协程来管理协程池
负责分配goroutine ,结束空闲的goroutine
所有的任务先提交到任务队列或者等待队列
如果有空闲的goroutine，分配使用，没有就新开一个
当新开一个协程就+1,如果已经到达最大数量，先放入等待队列中

杀死协程的方法
*/

// NewPool m是等待队列的最大任务数，n是最大协程数
func NewPool(m, n uint) *Pool {
	//n不能为0
	if n <= 0 {
		n = 1
	}
	pool := &Pool{
		countWorker: 0,
		maxWorker:   n,
		taskQueue:   make(chan func(v ...any)),
		waitQueue:   NewWait(m),
		workQueue:   make(chan func(v ...any)),
		status:      run,
		Logger:      log.New(os.Stdout, "[pool]"+time.Now().Format(time.DateTime)+" ", log.Llongfile),
	}
	//启动一个
	go pool.Star(context.Background())
	return pool
}

func (p *Pool) Submit(ctx context.Context, f func(v ...interface{})) error {
	if p.status == stop {
		return errors.New("pool池已经关闭！")
	}
	if f != nil {
		p.taskQueue <- f
		return nil
	}
	return errors.New("f func(v ...interface{}) 不能为空！")
}

func (p *Pool) Len() int {
	p.look.Lock()
	defer p.look.Unlock()
	return len(p.workQueue)
}

func (p *Pool) Star(ctx context.Context) {
	var wg = sync.WaitGroup{}
	timeout := time.NewTimer(time.Second * 3)

Loop:
	//	核心方法
	for {
		//等待队列里面不为空
		if len(p.waitQueue.channel) > 0 {
			//消费
			p.Println("等待队列取出")
			p.waitQueue.Cusmer(p.workQueue)
		}
		//	任务队列中
		select {
		case work, ok := <-p.taskQueue:
			p.Println("阻塞", ok)
			//不存在
			if !ok {
				break Loop
			}
			//存在
			select {
			//放进执行队列
			case p.workQueue <- work:
				fmt.Println("收入")
			//	没有协程来接收，新开一个
			default:
				//如果小于最大goroutine数目，新开一个

				if p.countWorker < p.maxWorker {
					p.Printf("当前协程数 %v", p.countWorker)
					wg.Add(1)
					go p.work(ctx, work, p.workQueue, &wg)
					p.countWorker++
				} else {
					//    送进等待队列
					p.Println("送入等待队列")
					ok := p.waitQueue.Push(work)
					if !ok {
						p.Println("waitQueue 容量过小！")
						break Loop
					}
				}

			}
			p.idle = false
		case <-timeout.C:
			fmt.Println("timeout")
			if p.idle && p.countWorker > 0 {
				select {
				//管道中会有一个goroutine接收到nil，退出
				case p.workQueue <- nil:
					p.countWorker--
					p.Println("当前协程池中协程数", p.countWorker)
				}
			} else if p.countWorker == 0 {
				break Loop
			}
			p.idle = true
			timeout.Reset(time.Second * 3)
		}
	}
	//这个时候是将所以goroutine全部退出后在退出协程
	p.Stop(context.Background())
	timeout.Stop()
	wg.Wait()
}

func (p *Pool) work(ctx context.Context, task func(...any), c chan func(v ...any), wg *sync.WaitGroup) {
	defer wg.Done()
	for task != nil {
		task()
		task = <-c
	}
}

func (p *Pool) Stop(ctx context.Context) {
	//关闭
	p.once.Do(func() {
		close(p.workQueue)
		close(p.waitQueue.channel)
		close(p.taskQueue)
		// workers.
		p.look.Lock()
		p.status = stop
		p.look.Unlock()
	})

}

func (p *Pool) stops() {
	for p.countWorker > 0 {
		p.workQueue <- nil
		p.countWorker--
	}
}

func (p *Pool) Wait() {
	for p.status != stop && p.countWorker != 0 {
	}
	p.Println("协程全部退出，关闭协程池")
}
