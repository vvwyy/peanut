package future

import (
	"context"
	"github.com/vvwyy/peanut/concurrent"
	"github.com/vvwyy/peanut/concurrent/executor"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	// NEW -> COMPLETING -> NORMAL
	// NEW -> COMPLETING -> EXCEPTIONAL
	// NEW -> CANCELLED
	// NEW -> INTERRUPTING -> INTERRUPTED
	NIL          = int32(-1)
	NEW          = int32(0)
	COMPLETING   = int32(1)
	NORMAL       = int32(2)
	ERROR        = int32(3)
	CANCELLED    = int32(4)
	INTERRUPTING = int32(5)
	INTERRUPTED  = int32(6)
)

type FutureTask struct {
	//parent context.Context
	//current context.Context

	mu         sync.Mutex // protects following fields
	state      int32
	executable executor.Executable
	waiters    *WaitNode
	//ctx        context.Context

	err    error
	result interface{}
}

func (futureTask *FutureTask) Run() {
	// todo set runner

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
	if futureTask.state != NEW {
		return false
	}
	newState := CANCELLED
	if mayInterruptIfRunning {
		newState = INTERRUPTING
	}
	if !atomic.CompareAndSwapInt32(&futureTask.state, NEW, newState) {
		return false
	}
	if mayInterruptIfRunning {
		// todo 中断当前任务
	}

	futureTask.finishCompletion()
	return true
}

func (futureTask *FutureTask) IsCancelled() bool {
	return futureTask.state >= CANCELLED
}

func (futureTask *FutureTask) IsDone() bool {
	return futureTask.state != NEW
}

func (futureTask *FutureTask) Get() (interface{}, error) {
	s := futureTask.state
	if s <= COMPLETING {
		ctx, cancelFunc := context.WithCancel(context.TODO())
		var err error
		s, err = futureTask.awaitDone(ctx, cancelFunc, false, 0)
		if err != nil {
			return nil, err
		}
	}
	return futureTask.report(s)
}

func (futureTask *FutureTask) GetWithTimeout(d time.Duration) (interface{}, error) {
	// todo
	return nil, nil
}

// ---------------------------------------------------------------------------------------------------------------------

func (futureTask *FutureTask) report(state int32) (interface{}, error) {
	ret := futureTask.result
	if state == NORMAL {
		return ret, nil
	}
	if state >= CANCELLED {
		return nil, concurrent.CancellationError
	}
	return nil, concurrent.ExecutionError
}

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
	// nothing to to currently
}

// Thinking: how to note goroutine,
// here we use context that binding to goroutine
type WaitNode struct {
	gotx   context.Context
	cancel context.CancelFunc
	next   *WaitNode
}

// Removes and signals all waiting goroutine
func (futureTask *FutureTask) finishCompletion() {
	for node := futureTask.waiters; node != nil; {
		p := unsafe.Pointer(futureTask.waiters)
		if atomic.CompareAndSwapPointer(&p, node, nil) {
			for {
				// signals waiting goroutine
				// todo
				nextNode := node.next
				if nextNode == nil {
					break
				}
				node.next = nil
				node = nextNode
			}
			break
		}
	}
	futureTask.done()
	futureTask.executable = nil
}

func (futureTask *FutureTask) done() {
	// nothing to do currently
}

// Awaits completion or aborts on interrupt or timeout.
func (futureTask *FutureTask) awaitDone(ctx context.Context, cancelFunc context.CancelFunc, timed bool, nanos time.Duration) (int32, error) {
	var deadline = time.Now()
	if timed {
		deadline = deadline.Add(nanos)
	}

	queued := false
	var q *WaitNode = nil
	for {
		if ctx.Err() != nil {
			return NIL, concurrent.InterruptedError
		}

		s := futureTask.state
		if s > COMPLETING {
			if q != nil {
				q.gotx = nil
			}
			return s, nil
		} else if s == COMPLETING {
			// need to yield
			time.Sleep(10 * time.Microsecond)
			continue
		} else if q == nil {
			q = &WaitNode{gotx: ctx, cancel: cancelFunc}
		} else if !queued {
			q.next = futureTask.waiters
			pointer := unsafe.Pointer(q.next)
			queued = atomic.CompareAndSwapPointer(&pointer, q.next, q)
		} else if timed {
			nanos = deadline.Sub(time.Now())
			if nanos <= 0 {
				futureTask.removeWaiter(q)
				return futureTask.state, nil
			}
			time.Sleep(nanos)
		} else {
			// park
			<-q.gotx.Done()
		}
	}
}

// Tries to unlink a timed-out or interrupted wait node
func (futureTask *FutureTask) removeWaiter(node *WaitNode) {
	if node == nil {
		return
	}
	node.gotx = nil
	// Thinking:
retry:
	for {
		var pred *WaitNode
		var s *WaitNode
		for q := futureTask.waiters; q != nil; q = s {
			s = q.next
			if q.gotx != nil {
				pred = q
			} else if pred != nil {
				pred.next = s
				if pred.gotx == nil { // check for race
					continue retry
				}
			} else {
				qPointer := unsafe.Pointer(q)
				if !atomic.CompareAndSwapPointer(&qPointer, q, s) {
					continue retry
				}
			}
		}
		break
	}
}
