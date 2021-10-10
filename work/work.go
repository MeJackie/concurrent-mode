/*
 * @Author: your name
 * @Date: 2021-10-10 14:59:55
 * @LastEditTime: 2021-10-10 23:16:14
 * @LastEditors: Please set LastEditors
 * @Description: 何使用无缓冲的通道来创建一个 goroutine 池，这些 goroutine 执行
 *               并控制一组工作，让其并发执行。在这种情况下，使用无缓冲的通道要比随意指定一个缓冲区大
 *               小的有缓冲的通道好，因为这个情况下既不需要一个工作队列，也不需要一组 goroutine 配合执
 *               行。无缓冲的通道保证两个 goroutine 之间的数据交换。这种使用无缓冲的通道的方法允许使用
 *               者知道什么时候 goroutine 池正在执行工作，而且如果池里的所有 goroutine 都忙，无法接受新的
 *               工作的时候，也能及时通过通道来通知调用者。使用无缓冲的通道不会有工作在队列里丢失或者
 *               卡住，所有工作都会被处理。
 * @FilePath: /concurrent-mode/work/work.go
 */
package work

import "sync"

type Worker interface {
	Task()
}

type Pool struct {
	work chan Worker
	wg   sync.WaitGroup
}

func New(MaxGoroutine int) *Pool {
	p := Pool{
		work: make(chan Worker),
	}

	p.wg.Add(MaxGoroutine)
	for i := 0; i < MaxGoroutine; i++ {
		go func() {
			// for range 循环会一直阻塞，直到从 work 通道收到一个 Worker 接
			// 口值。如果收到一个值，就会执行这个值的 Task 方法。一旦 work 通道被关闭，for range
			// 循环就会结束，并调用 WaitGroup 的 Done 方法。然后 goroutine 终止。
			for w := range p.work {
				w.Task()
			}
		}()
	}

	return &p
}

func (p *Pool) Run(fn Worker) {
	// work 无缓冲通道，堵塞直到有空闲gorotine
	p.work <- fn
}

func (p *Pool) Shutdown() {
	close(p.work)
	// 等待所有gorotine执行完
	p.wg.Wait()
}
