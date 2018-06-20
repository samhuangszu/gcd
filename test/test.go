package main

import (
	"fmt"
	"time"

	"github.com/samhuangszu/gcd"
)

func main() {

	gcd.Async(func(args ...interface{}) {
		fmt.Println(args)
	}, func(in int) (string, error) {
		return fmt.Sprintf("arg:%d", in), nil
	}, 90)

	gcd.Async(func(args ...interface{}) {
		fmt.Println(args)
	}, func(in string) (string, error) {
		return in, nil
	}, "test")

	gcd.AsyncTask(func(r gcd.TResult) {
		fmt.Println(r)
	}, func(args ...interface{}) gcd.Result {
		return gcd.Result{
			Error: nil,
			Data:  args,
		}
	}, "test")

	gcd.AsyncTask(func(r gcd.TResult) {
		fmt.Println(r)
	}, func(args ...interface{}) gcd.Result {
		return gcd.Result{
			Error: nil,
			Data:  args,
		}
	}, "test1", 1, "test2")

	gcd.AsyncTask(func(r gcd.TResult) {
		fmt.Println(r)
	}, func(args ...interface{}) gcd.Result {
		return gcd.Result{
			Error: nil,
			Data:  args,
		}
	}, "test3", 9, "test4")
	//可以试着注释这行代码，看看什么效果
	time.Sleep(time.Second * 10)

}
