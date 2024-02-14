package closer

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Closer struct {
	mu    sync.Mutex
	funcs []Func
}

func (c *Closer) Add(f Func) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.funcs = append(c.funcs, f)
}

func (c *Closer) Close(ctx context.Context) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var (
		complete = make(chan struct{}, 1)
	)

	go func() {
		for _, f := range c.funcs {
			if errF := f(ctx); errF != nil {
				err = errors.Join(err, errF)
			}
		}
		complete <- struct{}{}
	}()

	select {
	case <-complete:
		break
	case <-ctx.Done():
		return fmt.Errorf("shutdown cancelled: %v", ctx.Err())
	}

	return
}

type Func func(ctx context.Context) error
