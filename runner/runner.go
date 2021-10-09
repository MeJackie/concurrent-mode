/*
 * @Author: jackie
 * @Date: 2021-10-09 15:46:23
 * @LastEditTime: 2021-10-09 16:56:20
 * @LastEditors: Please set LastEditors
 * @Description: runner并发模型,适合后天运行任务，如cron等
 * @FilePath: /concurrent-mode/runner.go
 */
package runner

import (
	"errors"
	"os"
	"os/signal"
	"time"
)

type Runner struct {
	// 系统中断信号
	interrupt chan os.Signal

	// 超时信号
	timeout <-chan time.Time

	// 任务完成信号,成功时error为nil
	complete chan error

	// 任务
	tasks []func(int)
}

func New(d time.Duration) *Runner {
	return &Runner{
		//通道 interrupt 被初始化为缓冲区容量为 1 的通道。这可以保证通道至少能接收一个来自
		//语言运行时的 os.Signal 值，确保语言运行时发送这个事件的时候不会被阻塞。如果 goroutine
		//没有准备好接收这个值，这个值就会被丢弃。例如，如果用户反复敲 Ctrl+C 组合键，程序只会
		//在这个通道的缓冲区可用的时候接收事件，其余的所有事件都会被丢弃。
		interrupt: make(chan os.Signal, 1),
		timeout:   time.After(d),
		complete:  make(chan error), // 发送完成信号后，堵塞等待接受
	}
}

var ErrTimeout = errors.New("timeout")
var ErrInterrupted = errors.New("interrupted")

func (r *Runner) Start() error {
	// 注册中断信号接受者
	signal.Notify(r.interrupt, os.Interrupt)

	// 执行任务
	go func() {
		r.complete <- r.run()
	}()

	// 堵塞，检查超时、完成情况
	select {
	case err := <-r.complete:
		return err
	case <-r.timeout:
		return ErrTimeout
	}
}

func (r *Runner) Add(tasks ...func(int)) {
	r.tasks = append(r.tasks, tasks...)
}

func (r *Runner) run() error {
	for id, task := range r.tasks {
		if r.gotInterrupt() {
			return ErrInterrupted
		}
		task(id)
	}

	return nil
}

// gotInterrupt 验证是否接收到了中断信号
func (r *Runner) gotInterrupt() bool {
	select {
	case <-r.interrupt:
		signal.Stop(r.interrupt)
		return true
	default:
		return false
	}
}
