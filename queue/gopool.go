package queue

import (
	"context"
	"errors"
	"log"
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
	taskQueue chan func(...interface{})
	//waitQueue  等待队列
	waitQueue *Wait
	//	执行队列
	workQueue chan func(...interface{})
	//锁
	look sync.Mutex
	//	status 状态
	status int
	//执行一次
	once sync.Once
	//id
	idle bool
}

/*
先启动一个协程来管理协程池
负责分配goroutine ,结束空闲的goroutine
所有的任务先提交到任务队列或者等待队列
如果有空闲的goroutine，分配使用，没有就新开一个
当新开一个协程就+1,如果已经到达最大数量，先放入等待队列中

杀死协程的方法


*/

func (p *Pool) NewPool(m, n uint) *Pool {
	//n不能为0
	if n <= 0 {
		p.maxWorker = 1
	}
	pool := &Pool{
		countWorker: 0,
		maxWorker:   n,
		taskQueue:   make(chan func(v ...interface{})),
		waitQueue:   NewWait(m),
		workQueue:   make(chan func(v ...interface{})),
		status:      run,
	}
	//启动一个
	go p.Star(context.Background())
	return pool
}

func (p *Pool) Submit(ctx context.Context, f func(v ...interface{})) error {
	if p.status == stop {
		return errors.New("pool池已经关闭！")
	}
	if f != nil {
		p.taskQueue <- f
	}
	return errors.New("f func(v ...interface{}) 不能为空！")
}

func (p *Pool) Len() int {
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
			p.waitQueue.cusmer(p.workQueue)
		}
		//	任务队列中
		select {
		case work, ok := <-p.taskQueue:
			//不存在
			if !ok {
				break Loop
			}
			//存在,送入执行队列
			select {
			//新开一个goroutine
			case p.workQueue <- work:
				//如果小于最大goroutine数目，新开一个
				if p.countWorker < p.maxWorker {
					wg.Add(1)
					go p.work(work, p.workQueue, &wg)
					p.countWorker++
				}
			default:
				//    送进等待队列
				ok := p.waitQueue.push(work)
				if !ok {
					log.Println("waitQueue 容量过小！")
					break Loop
				}
			}
			p.idle = false
		case <-timeout.C:
			if p.idle && p.countWorker > 0 {
				select {
				case p.workQueue <- nil:
					p.countWorker--
				default:
					break
				}
			}
			p.idle = true
			timeout.Reset(time.Second * 3)
		}
	}

	p.Stop(context.Background())
	timeout.Stop()
	wg.Wait()
}

func (p *Pool) work(task func(...interface{}), c chan func(v ...interface{}), wg *sync.WaitGroup) {
	for c != nil {
		task()
		task = <-c
	}
	wg.Done()
}

func (p *Pool) Stop(ctx context.Context) {
	//关闭
	p.once.Do(func() {
		p.stops()
		close(p.workQueue)
		close(p.waitQueue.channel)
		close(p.taskQueue)
		// workers.
		p.look.Lock()
		p.status = stop
		p.look.Unlock()
		close(p.taskQueue)
	})

}

func (p *Pool) stops() {
	for p.countWorker > 0 {
		p.workQueue <- nil
		p.countWorker--
	}
}
