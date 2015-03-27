// Copyright 2015 Mathieu MAST. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package concurrency

import (
	"errors"
	"github.com/goloc/container"
	"sync"
	"time"
)

type Future struct {
	promise *Promise
}

func NewFuture() *Future {
	f := new(Future)
	f.promise = new(Promise)
	f.promise.dones = container.NewLinkedList()
	f.promise.fails = container.NewLinkedList()
	f.promise.ends = container.NewLinkedList()
	return f
}

func (f *Future) Resolve(element interface{}) {
	f.promise.done(element)
}

func (f *Future) Reject(err error) {
	f.promise.fail(err)
}

func (f *Future) GetPromise() *Promise {
	return f.promise
}

type Promise struct {
	closed  bool
	dones   *container.LinkedList
	fails   *container.LinkedList
	ends    *container.LinkedList
	element interface{}
	err     error
	mutex   sync.Mutex
}

func (p *Promise) close() {
	p.closed = true
	p.dones = nil
	p.fails = nil
	p.ends.Visit(func(ch interface{}, i int) {
		ch.(chan bool) <- true
	})
}

func (p *Promise) done(element interface{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.closed {
		return
	}
	p.element = element
	p.dones.Visit(func(f interface{}, i int) {
		go f.(func(interface{}))(p.element)
	})
	p.close()
}

func (p *Promise) fail(err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.closed {
		return
	}
	p.err = err
	p.fails.Visit(func(f interface{}, i int) {
		go f.(func(interface{}))(p.err)
	})
	p.close()
}

func (p *Promise) Then(done func(interface{}), fail func(error)) *Promise {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.closed {
		if p.err != nil {
			go fail(p.err)
		} else {
			go done(p.element)
		}
	} else {
		p.dones.Add(done)
		p.fails.Add(fail)
	}
	return p
}

func (p *Promise) Wait(mawWait time.Duration) (interface{}, error) {
	if p.closed {
		return p.element, p.err
	}
	end := make(chan bool, 1)
	p.ends.Add(end)
	select {
	case <-end:
		return p.element, p.err
	case <-time.After(mawWait):
		return nil, errors.New("Timeout")
	}
}
