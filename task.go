package gcd

import (
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/satori/go.uuid"
)

// Taskfunc 任务方法
type Taskfunc func(args ...interface{}) Result

// Replyfunc 回调方法
type Replyfunc func(r TResult)

// TResult 任务结果
type TResult struct {
	ID       string        //任务ID
	TaskType int           //任务类型
	Args     []interface{} //任务参数
	Result   Result        //执行结果
}

// Result 结果集
type Result struct {
	Error error       //错误信息
	Data  interface{} //返回结果
}

// TTask 任务
type TTask struct {
	id       string
	args     []interface{}
	taskType int
	task     Taskfunc
}

// TPool 任务池
type TPool struct {
	graceful    chan byte      //退出任务
	gSugNum     int32          //初始goroutine数量
	gCurNum     int32          //当前goroutine数量
	totalTask   int32          //总任务量
	totalResult int32          //总的结果数
	taskQueue   chan TTask     //任务队列
	resultQueue chan TResult   //任务执行的结果队列
	gchans      []chan byte    //管理goroutine数量
	wg          sync.WaitGroup //等待任务执行
}

// 常量
const (
	TaskQueueMaxSize    = 100000 //任务队列
	ResultQuequeMaxSize = 2
	MaxGoRoutineNum     = 100
	MinGoRoutineNum     = 5
	MonitorInterval     = 1
	ScaleCheckMax       = 5
)

var logger = log.New(os.Stdout, "debug: ", log.LstdFlags|log.Lshortfile)

// NewTPool 生成对象
func NewTPool(gSugNum int32) *TPool {
	taskQueue := make(chan TTask, TaskQueueMaxSize)
	resultQueue := make(chan TResult, ResultQuequeMaxSize)
	gracechan := make(chan byte)
	if gSugNum < MinGoRoutineNum {
		gSugNum = MinGoRoutineNum
	}
	return &TPool{gSugNum: gSugNum, taskQueue: taskQueue, resultQueue: resultQueue, graceful: gracechan}
}

// AddTask 添加任务
func (pool *TPool) AddTask(task Taskfunc, taskType int, args ...interface{}) string {
	uuid, err := uuid.NewV4()
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	if err == nil {
		id = uuid.String()
	}
	//logger.Printf("add %v start\n", id)
	pool.taskQueue <- TTask{id: id, task: task, taskType: taskType, args: args}
	//logger.Printf("add %v stop\n", id)
	return id
}

// 检查任务量
func (pool *TPool) scale() {
	timer := time.NewTicker(MonitorInterval * time.Second)
	var scale int16

	for {
		select {
		case <-timer.C:
			// logger.Printf("gCurNum: %v\ttotal tasks: %v\n", pool.gCurNum, len(pool.taskQueue))
			switch {
			case len(pool.taskQueue) >= int(2*pool.gCurNum):
				if pool.gCurNum*2 <= MaxGoRoutineNum && scale >= ScaleCheckMax {
					//pool.gCurNum *= 2
					pool.incr(pool.gCurNum)
					scale = 0
				} else if scale < ScaleCheckMax {
					scale++
				}

			case len(pool.taskQueue) <= int(pool.gCurNum/2):
				if pool.gCurNum/2 >= MinGoRoutineNum && scale <= -1*ScaleCheckMax {
					// logger.Printf("desc action - gNum: %v\n", pool.gCurNum)
					pool.desc(pool.gCurNum / 2)
					scale = 0
				} else if scale > -1*ScaleCheckMax {
					scale--
				}

			default:
				fmt.Println("no scale!")
			}
		case <-pool.graceful:
			{
				fmt.Println("graceful exit!")
				return
			}
		}
		// logger.Printf("pool Goroutine num: %v\tscale: %v\n", pool.gCurNum, scale)
	}
}

// 减少goroutine
func (pool *TPool) desc(num int32) {
	// logger.Printf("desc info - num: %v\t gchansLen: %v\n", num, len(pool.gchans))
	if int(num) > len(pool.gchans) {
		return
	}
	for _, c := range pool.gchans[:len(pool.gchans)-int(num)] {
		c <- 1
	}
	pool.gchans = pool.gchans[len(pool.gchans)-int(num):]
}

// 添加goroutine
func (pool *TPool) incr(num int32) {
	for i := 0; i < int(num); i++ {
		//logger.Printf("!add %v goroutine\n", i)
		ch := make(chan byte)
		pool.gchans = append(pool.gchans, ch)
		//id := uuid.NewV4().String()
		pool.wg.Add(1)
		go func(mych <-chan byte) {
			atomic.AddInt32(&(pool.gCurNum), 1)
			//logger.Printf("the goroutine %v\n", mych)
			defer func() {
				fmt.Println("closed")
				atomic.AddInt32(&pool.gCurNum, -1)
				pool.wg.Done()
			}()

		ExitLoop:
			for {
				select {
				case task, ok := <-pool.taskQueue:
					if ok {
						res := task.task(task.args...)
						pool.resultQueue <- TResult{ID: task.id, Result: res, TaskType: task.taskType, Args: task.args}
					} else {
						break ExitLoop
					}
				case cmd := <-mych:
					switch cmd {
					case 1:
						break ExitLoop
					}
				case <-pool.graceful:
					{
						break ExitLoop
					}
				}
			}
		}(ch)
	}
}

// Start 开始任务初始goroutine
func (pool *TPool) Start() {
	pool.incr(pool.gSugNum)
	go pool.scale()
}

// Stop 停止
func (pool *TPool) Stop() {
	close(pool.taskQueue)
	pool.wg.Wait()
	close(pool.resultQueue)
	close(pool.graceful)
}

// GetResult 监听结果获取
func (pool *TPool) GetResult(f func(res TResult)) {
	go func() {
		for r := range pool.resultQueue {
			f(r)
			atomic.AddInt32(&pool.totalResult, 1)
			atomic.AddInt32(&pool.totalTask, 1)
		}
	}()
}

// Heatbeat 心跳log
func (pool *TPool) Heatbeat() {
	go func() {
		timer := time.NewTicker(5 * time.Second)
		defer func() {
			timer.Stop()
		}()
	ExitLoop:
		for {
			select {
			case <-timer.C:
				{
					logger.Printf("gCurNum: %v, taskQueue: %v, resultQueue: %v totalResult: %v\n", pool.gCurNum, len(pool.taskQueue), len(pool.resultQueue), pool.totalResult)
				}
			case <-pool.graceful:
				{
					fmt.Println("graceful exit!")
					break ExitLoop
				}
			}
		}
	}()
}
