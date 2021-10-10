/*
 * @Author: jackie
 * @Date: 2021-10-09 18:45:41
 * @LastEditTime: 2021-10-10 22:29:13
 * @LastEditors: Please set LastEditors
 * @Description: 资源池模式
 *               使用有缓冲的通道实现资源池，来管理可以在
 *               任意数量的goroutine之间共享及独立使用的资源。这种模式在需要共享一组静态资源的情况（如
 *               共享数据库连接或者内存缓冲区）下非 常有用。如果goroutine需要从池里得到这些资源中的一个，
 *               它可以从池里申请，使用完后归还到资源池里
 * @FilePath: /concurrent-mode/pool/pool.go
 */
package pool

import (
	"errors"
	"io"
	"log"
	"sync"
)

type Pool struct {
	m         sync.Mutex
	resources chan io.Closer
	factory   func() (io.Closer, error)
	closed    bool
}

var ErrPoolClosed = errors.New("pool is closed")

func New(fn func() (io.Closer, error), size uint) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("Size value too small.")
	}

	return &Pool{
		resources: make(chan io.Closer, size),
		factory:   fn,
	}, nil
}

func (p *Pool) Acquire() (io.Closer, error) {
	select {
	case r, ok := <-p.resources:
		log.Println("Acquire:", "Share Resource")
		if !ok {
			return nil, ErrPoolClosed
		}
		return r, nil
	default:
		// 资源队列为空
		log.Println("Acquire:", "New Resource")
		return p.factory()
	}
}

// 使用互斥量有两个目的
// 第一，可以保护第 65 行中读取 closed
// 标志的行为，保证同一时刻不会有其他 goroutine 调用 Close 方法写同一个标志。第二，我们不
// 想往一个已经关闭的通道里发送数据，因为那样会引起崩溃。如果 closed 标志是 true，我们
// 就知道 resources 通道已经被关闭。
func (p *Pool) Release(r io.Closer) {
	// 保证本操作与Close的操作安全
	p.m.Lock()
	defer p.m.Unlock()
	if p.closed {
		r.Close()
		return
	}

	select {
	case p.resources <- r:
		log.Println("Release:", "In Queue")
	default:
		// 队列已满或已经关闭通道
		log.Println("Release:", "Closing")
		r.Close()
	}
}

func (p *Pool) Close() {
	// 保证本操作与Release的操作安全
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return
	}

	p.closed = true

	// todo ?? 必须先关闭通道，否则回和r.Close发生死锁
	close(p.resources)

	for r := range p.resources {
		r.Close()
	}
}
