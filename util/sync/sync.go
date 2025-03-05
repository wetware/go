package syncutils

import (
	"sync"

	"go.uber.org/atomic"
)

type Any struct {
	wg sync.WaitGroup
	er atomic.Error
}

func (a *Any) Wait() error {
	a.wg.Wait()
	return a.er.Load()
}

func (a *Any) Go(fn func() error) {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.er.CompareAndSwap(nil, fn())
	}()
}
