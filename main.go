package main

import (
	"context"
	"fmt"
	"github.com/Xiaoxusheng/pool/queue"
	"log"
	"strconv"
	"time"
)

func main() {
	pool := queue.NewPool(20, 10)
	ctx := context.Background()

	for i := 0; i < 300; i++ {
		i = i
		err := pool.Submit(ctx, func(v ...any) {
			log.Println(i)
			add(1, 2, i)
		})
		if err != nil {
			log.Println(err)
			return
		}
	}
	pool.Wait()
}

func add(a, b, i int) {
	time.Sleep(time.Second)
	fmt.Println("第"+strconv.Itoa(i)+"个任务执行完成", a+b)
}
