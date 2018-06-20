package gcd

import (
	"reflect"

	"github.com/aurelien-rainone/assertgo"
)

// Async 把方法转成异步
func Async(fn interface{}, handler interface{}, params ...interface{}) {
	assert.True(isFun(fn), "fn is't func")
	assert.Truef(isFun(handler), "handler should be fun:%v", handler)
	ch := make(chan byte, 1)
	values := make([]reflect.Value, 0)
	go func(fn interface{}, handler interface{}, ch chan<- byte, values *[]reflect.Value, params ...interface{}) {
		handlerFun := reflect.ValueOf(handler)
		paramNum := len(params)
		paramValues := make([]reflect.Value, paramNum)
		if paramNum > 0 {
			for k, v := range params {
				paramValues[k] = reflect.ValueOf(v)
			}
		}
		//调用结束
		*values = handlerFun.Call(paramValues)
		close(ch)
	}(fn, handler, ch, &values, params...)
	select {
	case <-ch:
		{
			fnFun := reflect.ValueOf(fn)
			fnFun.Call(values)
			return
		}
	}
}

func isFun(fn interface{}) bool {
	if fn == nil {
		return false
	}
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() == reflect.Func {
		return true
	}
	return false
}
