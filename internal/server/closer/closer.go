package closer

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Func shutdown function for Closer
type Func func(ctx context.Context) error

// Closer for graceful shutdown
type Closer struct {
	funcs []Func
	names []string
	mu    sync.Mutex
}

// Add shutdown service function
func (c *Closer) Add(n string, f Func) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.names = append(c.names, n)
	c.funcs = append(c.funcs, f)
}

// Close initialize graceful shutdown - run all shutdwon functions
func (c *Closer) Close(ctx context.Context) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var (
		wg       sync.WaitGroup
		complete = make(chan struct{}, 1)
	)

	wg.Add(len(c.funcs))
	for i, f := range c.funcs {
		i, f := i, f
		go func() {
			if errF := f(ctx); errF != nil {
				err = errors.Join(err, fmt.Errorf("close %s error: %w", c.names[i], errF))
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		complete <- struct{}{}
	}()

	select {
	case <-complete:
		break
	case <-ctx.Done():
		return fmt.Errorf("shutdown cancelled: %s", ctx.Err())
	}

	return
}
