# AsyncTask
类似oc block,回调都放在一个代码流里

##使用方式使用方式：

1.第一个参数是TaskFunc类型的方法，负责执行任务并返回gcd.Result结果

2.第二个参数是Replyfunc类型的方法，任务执行完后，回调这个方法，把结果封装成gcd.TResult

3.第三个参数是args...interface{},taskFunc的参数，同时存在gcd.TResult的args中

```go
    gcd.AsyncTask(func(r gcd.TResult) {
      fmt.Println(r)
    }, func(args ...interface{}) gcd.Result {
      return gcd.Result{
        Error: nil,
        Data:  args,
      }
    }, "test")

```

*这份代码中的task.go 基于一份代码修改，一时找不到对应的链接，有知道的朋友请知悉*

# Async
把一个方法转成异步调用，并最终返回当前协程处理结果

##使用方法

1.Async 的第三个以后的参数，对应每二个fun的参数，必须一致

```go
  gcd.Async(func(args ...interface{}) {
		c.Output(args)
	}, func(test int) (string, error) {
		start := time.Now().UnixNano()
		time.Sleep(time.Second * 10)
		end := time.Now().UnixNano()
		return fmt.Sprintf("arg:%d\nstart:%d\nend:%d\n", test, start, end), nil
	}, 90)
```


# Async 与AsyncTask 的区别

  1.Async 会回调到当前routine处理最终结果，所以当前routine会一直等待reply处理完成再退出

  2.AsyncTask 回调方法会新起一个routine处理，所以当前routine执行完就马上退出；
