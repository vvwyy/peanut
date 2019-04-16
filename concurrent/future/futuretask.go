package future

import (
	"context"
	"github.com/vvwyy/peanut/concurrent/executor"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// NEW -> COMPLETING -> NORMAL
	// NEW -> COMPLETING -> EXCEPTIONAL
	// NEW -> CANCELLED
	// NEW -> INTERRUPTING -> INTERRUPTED
	NEW               = int32(0)
	COMPLETING        = int32(1)
	NORMAL            = int32(2)
	ERROR             = int32(3)
	STATE_CANCELLED   = int32(4)
	INTERRUPTING      = int32(5)
	STATE_INTERRUPTED = int32(6)
)

type FutureTask struct {
	//parent context.Context
	//current context.Context

	mu         sync.Mutex // protects following fields
	state      int32
	executable executor.Executable
	ctx        context.Context

	err    error
	result interface{}
}

func (futureTask *FutureTask) Run() {
	e := futureTask.executable
	if e == nil || futureTask.state != NEW {
		return
	}
	result, err := e.Run()
	if err != nil {
		futureTask.setError(err)
	}
	futureTask.setResult(result)

	state := futureTask.state
	if state >= INTERRUPTING {
		futureTask.handlePossibleCancellationInterrupt(state)
	}
}

func (futureTask *FutureTask) Cancel(mayInterruptIfRunning bool) bool {
	// todo
	return false
}

func (futureTask *FutureTask) IsCancelled() bool {
	// todo
	return false
}

func (futureTask *FutureTask) IsDone() bool {
	// todo
	return false
}

func (futureTask *FutureTask) Get() (interface{}, error) {
	// todo
	return nil, nil
}

func (futureTask *FutureTask) GetWithTimeout(d time.Duration) (interface{}, error) {
	// todo
	return nil, nil
}

// ---------------------------------------------------------------------------------------------------------------------

func (futureTask *FutureTask) setError(err error) {
	if atomic.CompareAndSwapInt32(&futureTask.state, NEW, COMPLETING) {
		futureTask.mu.Lock()
		futureTask.err = err
		futureTask.state = ERROR
		futureTask.result = nil
		futureTask.finishCompletion()
		futureTask.mu.Unlock()
	}
}

func (futureTask *FutureTask) setResult(ret interface{}) {
	if atomic.CompareAndSwapInt32(&futureTask.state, NEW, COMPLETING) {
		futureTask.mu.Lock()
		futureTask.err = nil
		futureTask.state = NORMAL
		futureTask.result = ret
		futureTask.finishCompletion()
		futureTask.mu.Unlock()
	}
}

func (futureTask *FutureTask) handlePossibleCancellationInterrupt(state int32) {
	// nothing to to
}

func (futureTask *FutureTask) finishCompletion() {


}
