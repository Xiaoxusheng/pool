package _interface

import (
	"context"
)

type Pools interface {
	//Len  任务队列长度
	Len() int
	//Submit 提交任务
	Submit(ctx context.Context, w func(v ...interface{})) error
	// Stop   停止协程池
	Stop(ctx context.Context)
	//Star 启动
	Star(ctx context.Context)
	//
}
