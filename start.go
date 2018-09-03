package gcd

import (
	"sync"
)

// Task 任务池
type Task struct {
	taskPool  *TPool //任务池
	replyLock *sync.Mutex
	replyFunc map[string]Replyfunc
}

var gcdTask *Task
var once sync.Once

// GetTask 获取充电任务
func GetTask() *Task {
	once.Do(func() {
		gcdTask = initTask()
	})
	return gcdTask
}

func initTask() *Task {
	gcdTask := &Task{
		taskPool:  NewTPool(5),
		replyLock: new(sync.Mutex),
		replyFunc: make(map[string]Replyfunc, 0),
	}
	gcdTask.Start()
	gcdTask.gcdResult()
	return gcdTask
}

// AsyncTask 添加充电任务
func AsyncTask(replyFunc Replyfunc, taskFunc Taskfunc, args ...interface{}) {
	task := GetTask()
	task.AddTask(taskFunc, replyFunc, 0, args...)
}

// QuequeTask 只是添加到队列
func QuequeTask(replyFunc Replyfunc, args ...interface{}) {
	task := GetTask()
	task.AddTask(func(args ...interface{}) Result {
		return Result{
			Error: nil,
			Data:  nil,
		}
	}, replyFunc, 0, args...)
}

// AddTask 添加任务
func (c *Task) AddTask(taskFunc Taskfunc, replyFunc Replyfunc, taskType int, args ...interface{}) {
	id := c.taskPool.AddTask(taskFunc, taskType, args...)
	c.replyFunc[id] = replyFunc
}

func (c *Task) gcdResult() {
	c.taskPool.GetResult(func(r TResult) {
		if fn, ok := c.replyFunc[r.ID]; ok {
			fn(r)
			c.replyLock.Lock()
			delete(c.replyFunc, r.ID)
			c.replyLock.Unlock()
		}
	})
}

// Start 开启任务
func (c *Task) Start() {
	c.taskPool.Start()
}

// Stop 结束任务
func (c *Task) Stop() {
	c.taskPool.Stop()
}
