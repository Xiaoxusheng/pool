package main

import (
	"context"
	"fmt"
	"log"
	"pool/queue"
	"strconv"
	"time"
)

func main() {
	pool := queue.NewPool(20, 10)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		i = i
		err := pool.Submit(ctx, func(v ...interface{}) {
			log.Println(i)
			add(1, 2, i)
		})
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func add(a, b, i int) {
	time.Sleep(time.Second * 2)
	fmt.Println("第"+strconv.Itoa(i)+"个任务执行完成", a+b)
}
