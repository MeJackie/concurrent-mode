/*
 * @Author: your name
 * @Date: 2021-10-10 13:30:16
 * @LastEditTime: 2021-10-26 17:30:17
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: /concurrent-mode/main.go
 */
package main

import (
	"concurrent-mode/runner"
	"log"
	"os"
	"time"
)

func main() {
	
	testRunner()
}

const timeout = time.Second * 3

func testRunner() {
	log.Printf("处理开始")
	r := runner.New(timeout)
	r.Add(createTask(), createTask(), createTask())
	// 执行任务并处理结果
	if err := r.Start(); err != nil {
		switch err {
		case runner.ErrTimeout:
			log.Printf("任务执行超时")
			os.Exit(1)
		case runner.ErrInterrupted:
			log.Printf("任务执行中断")
			os.Exit(2)
		}
	}
	log.Printf("处理完成")
}

func createTask() func(int) {
	return func(id int) {
		log.Printf("执行任务 #%d.", id)
		time.Sleep(time.Duration(id) * time.Second)
	}
}
