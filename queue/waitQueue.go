package queue

// 等待队列
type Wait struct {
	Id      int
	channel chan func(...interface{})
}

// 消费等待队列方法
func (w *Wait) push(c func(v ...interface{})) bool {
	if len(w.channel) >= cap(w.channel) {
		return false
	}
	//	将消息推入
	if c != nil {
		w.channel <- c
	}
	return true

}

// 消费等待队列中的消息
func (w *Wait) cusmer(c chan func(v ...interface{})) {
	var t bool
	for !t {
		//考录任务队列容纳问题
		//当任务队列c中阻塞放不下任务时，退出循环
		select {
		case task, ok := <-w.channel:
			//ok为nil是是推出的信号，关闭等待队列
			if !ok {
				close(w.channel)
			}
			c <- task
		default:
			t = true
		}
	}

}

func NewWait(m uint) *Wait {
	return &Wait{
		Id:      0,
		channel: make(chan func(...interface{}), m),
	}
}
