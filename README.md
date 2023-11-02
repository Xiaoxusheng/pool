## go协程池简单实现



#### 1 因为自已需要一直写一个协程池来控制goroutine数量与并发情况


#### 2 目前正在写一个go协程池的简单实现，目前只实现了简单的功能，后续会继续完善


#### 3 用法

```go
package main

import (
	"context"
	"log"
	"pool/queue"
)

func main() {

	pool := queue.NewPool(20, 10)
	ctx := context.Background()

	for i := 0; i < 300; i++ {
		i = i
		err := pool.Submit(ctx, func(v ...interface{}) {
			log.Println(i)
		})
		if err != nil {
			log.Println(err)
			return
		}
	}
	pool.Wait()
}



```
