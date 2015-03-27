// Copyright 2015 Mathieu MAST. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package concurrency

import (
	"errors"
	"time"
)

type Semaphore struct {
	promise *Promise
	pool    chan bool
}

func NewSemaphore(permit int) *Semaphore {
	s := new(Semaphore)
	s.pool = make(chan bool, permit)
	for i := 0; i < permit; i++ {
		s.pool <- true
	}
	return s
}

func (s *Semaphore) Acquire(mawWait time.Duration) error {
	select {
	case <-s.pool:
		return nil
	case <-time.After(mawWait):
		return errors.New("Timeout")
	}
}

func (s *Semaphore) Release() {
	s.pool <- true
}
