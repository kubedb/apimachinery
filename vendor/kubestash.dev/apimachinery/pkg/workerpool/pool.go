/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package workerpool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type WorkerPool struct {
	ctx    context.Context
	cancel context.CancelFunc

	workers chan struct{}
	wg      sync.WaitGroup

	err          error
	completed    int64
	panicHandler func(any)
	mu           sync.Mutex
	firstErrOnce sync.Once
	timeout      time.Duration
}

// NewWorkerPool creates a WorkerPool with optional args:
//   - int: max concurrent workers
//   - time.Duration: timeout duration
//   - func(any): panic handler
func NewWorkerPool(ctx context.Context, args ...any) *WorkerPool {
	cctx, cancel := context.WithCancel(ctx)
	tp := &WorkerPool{
		ctx:          cctx,
		cancel:       cancel,
		workers:      make(chan struct{}, 1),
		panicHandler: func(p any) { fmt.Println("panic in WorkerPool:", p) },
	}
	tp.injectArgs(args...)
	return tp
}

func (wp *WorkerPool) injectArgs(args ...any) {
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			wp.workers = make(chan struct{}, v)
		case int64:
			wp.workers = make(chan struct{}, v)
		case time.Duration:
			cctx, cancel := context.WithTimeout(wp.ctx, v)
			wp.ctx, wp.cancel, wp.timeout = cctx, cancel, v
		case func(any):
			wp.panicHandler = v
		default:
			panic(fmt.Sprintf("unsupported argument type: %T", v))
		}
	}
}

func (wp *WorkerPool) Run(f func() error) {
	select {
	case wp.workers <- struct{}{}:
	case <-wp.ctx.Done():
		return
	}

	wp.wg.Add(1)
	go func() {
		defer func() {
			if p := recover(); p != nil && wp.panicHandler != nil {
				wp.panicHandler(p)
			}
			wp.done()
		}()

		if err := f(); err != nil {
			wp.setError(err)
			if wp.cancel != nil {
				wp.cancel()
			}
		}
	}()
}

func (wp *WorkerPool) Wait() error {
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// all tasks finished normally
	case <-wp.ctx.Done():
		// timeout or cancel triggered
		wp.setError(wp.ctx.Err())
	}

	return wp.err
}

func (wp *WorkerPool) done() {
	<-wp.workers
	atomic.AddInt64(&wp.completed, 1)
	wp.wg.Done()
}

func (wp *WorkerPool) Cancel() {
	if wp.cancel != nil {
		wp.cancel()
	}
}

func (wp *WorkerPool) setError(err error) {
	wp.firstErrOnce.Do(func() {
		wp.mu.Lock()
		wp.err = err
		wp.mu.Unlock()
	})
}

func (wp *WorkerPool) Completed() int64 {
	return atomic.LoadInt64(&wp.completed)
}
