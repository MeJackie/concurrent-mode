/*
 * @Author: jackie
 * @Date: 2021-10-09 18:45:41
 * @LastEditTime: 2021-10-09 19:03:53
 * @LastEditors: Please set LastEditors
 * @Description: 资源池模式
 * @FilePath: /concurrent-mode/pool/pool.go
 */
package pool

import (
	"errors"
	"io"
	"sync"
)

type Pool struct {
	m         sync.Mutex
	resources chan io.Closer
	factory   func() (io.Closer, error)
	closed    bool
}

var ErrPoolClosed = errors.New("pool is closed")

func New(fn func() (io.Closer, error), size int) (*Pool, error) {
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
		if !ok {
			return nil, ErrPoolClosed
		}
		return r, nil
	default:
		// 资源队列为空
		return p.factory()
	}
}

func (p *Pool) Release(r io.Closer) bool {
	// 保证本操作与Close的操作安全
	p.m.Lock()
	defer p.m.Unlock()
	if p.closed {
		r.Close()
		return true
	}

	select {
	case p.resources <- r:
		return true
	default:
		// 队列已满或已经关闭通道
		r.Close()
		return true
	}
}

func (p *Pool) Close() error {
	// 保证本操作与Release的操作安全
	p.m.Lock()
	defer p.m.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// todo ?? 必须先关闭通道，否则回和r.Close发生死锁
	close(p.resources)

	for r := range p.resources {
		r.Close()
	}
	return nil
}
