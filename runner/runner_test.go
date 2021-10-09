/*
 * @Author: jackie
 * @Date: 2021-10-09 15:46:23
 * @LastEditTime: 2021-10-09 18:13:33
 * @LastEditors: Please set LastEditors
 * @Description: runner并发模型,适合后天运行任务，如cron等
 * @FilePath: /concurrent-mode/runner.go
 */
package runner

import (
	"errors"
	"log"
	"os"
	"syscall"
	"testing"
	"time"
)

const timeout = time.Second * 4

func TestRunnerTimeout(t *testing.T) {
	r := New(timeout)
	r.Add(createTask(), createTask(), createTask(), createTask())
	t.Run("测试超时", func(t *testing.T) {
		if got := r.Start(); !errors.Is(got, ErrTimeout) {
			t.Errorf("Start() = %v, want %v", got, ErrTimeout)
		}
	})
}

func TestRunnerInterrupt(t *testing.T) {
	r := New(timeout)
	r.Add(createTask(), createTask(), createTask(), createTask())
	produceInterrupt := func() {
		<-time.After(1 * time.Second)
		t.Logf("主动生成中断信号")
		p, _ := os.FindProcess(syscall.Getpid())
		p.Signal(os.Interrupt)
	}
	go produceInterrupt()

	t.Run("测试中断", func(t *testing.T) {
		if got := r.Start(); !errors.Is(got, ErrInterrupted) {
			t.Errorf("Start() = %v, want %v", got, ErrInterrupted)
		}
	})
}

func createTask() func(int) {
	return func(id int) {
		log.Printf("执行任务 #%d.", id)
		time.Sleep(time.Duration(id) * time.Second)
	}
}
