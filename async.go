package gcd

import (
	"reflect"

	"github.com/aurelien-rainone/assertgo"
)

// Async 把方法转成异步
func Async(fn interface{}, handler interface{}, params ...interface{}) {
	assert.True(isFun(fn), "fn is't func")
	assert.Truef(isFun(handler), "handler should be fun:%v", handler)
	graceful := make(chan byte, 1) //退出routine
	values := make([]reflect.Value, 0)
	go func(fn interface{}, handler interface{}, graceful chan<- byte, values *[]reflect.Value, params ...interface{}) {
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
		close(graceful)
	}(fn, handler, graceful, &values, params...)
	select {
	case <-graceful:
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

func sliceInsert(s []interface{}, index int, value interface{}) []interface{} {
	rear := append([]interface{}{}, s[index:]...)
	return append(append(s[:index], value), rear...)
}

func sliceRemove(s []interface{}, index int) []interface{} {
	return append(s[:index], s[index+1:]...)
}
