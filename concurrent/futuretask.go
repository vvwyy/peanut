package concurrent

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	// NEW -> COMPLETING -> NORMAL
	// NEW -> COMPLETING -> ERROR
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
	mu           sync.Mutex // protects following fields
	state        int32
	executable   Executable
	waiters      *WaitNode
	runnerCtx    context.Context
	runnerCancel context.CancelFunc

	err    error
	result interface{}
}

func NewFutureTask(parentCtx context.Context, executable Executable) *FutureTask {
	f := &FutureTask{
		executable: executable,
		state:      NEW,
	}

	// set runner
	// not save parent context in future
	f.runnerCtx, f.runnerCancel = context.WithCancel(parentCtx)
	return f
}

func (futureTask *FutureTask) Run() {
	if futureTask.runnerCtx == nil {
		return
	}

	e := futureTask.executable
	if e == nil || futureTask.state != NEW {
		return
	}

	c := make(chan error, 1)

	go func() {
		result, err := e()
		if err != nil {
			futureTask.setError(err)
			c <- err
		} else {
			futureTask.setResult(result)
			c <- nil
		}
	}()

	select {
	case <-futureTask.runnerCtx.Done():
		futureTask.setError(futureTask.runnerCtx.Err())
	case err := <-c:
		if err != nil {
			futureTask.setError(err)
		}
	}


//ret:
//	for {
//		select {
//		case <-futureTask.runnerCtx.Done():
//			futureTask.setError(futureTask.runnerCtx.Err())
//			break ret
//		case err := <-c:
//			if err != nil {
//				futureTask.setError(err)
//			}
//		default:
//			if futureTask.IsDone() {
//				break ret
//			}
//		}
//	}

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
		// interrupt current task
		// Thinking: actually ctx cancel is not work well
		futureTask.runnerCancel()

		futureTask.mu.Lock()
		futureTask.state = INTERRUPTED
		futureTask.result = nil
		futureTask.err = CancellationError
		futureTask.finishCompletion()
		futureTask.mu.Unlock()
	}
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
		var err error
		s, err = futureTask.awaitDone(false, 0)
		if err != nil {
			return nil, err
		}
	}
	return futureTask.report(s)
}

func (futureTask *FutureTask) GetWithTimeout(d time.Duration) (interface{}, error) {
	s := futureTask.state
	if s <= COMPLETING {
		var err error
		s, err = futureTask.awaitDone(true, d)
		if err != nil {
			return nil, err
		}
		if s <= COMPLETING {
			return nil, TimeoutError
		}
	}
	return futureTask.report(s)
}

// ---------------------------------------------------------------------------------------------------------------------

func (futureTask *FutureTask) createWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(futureTask.runnerCtx)
}

func (futureTask *FutureTask) report(state int32) (interface{}, error) {
	ret := futureTask.result
	if state == NORMAL {
		return ret, nil
	}
	if state >= CANCELLED {
		return nil, CancellationError
	}
	return nil, ExecutionError
}

func (futureTask *FutureTask) setError(err error) {
	if err == nil {
		return
	}
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
	if ret == nil {
		return
	}
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
		if atomic.CompareAndSwapUintptr((*uintptr)(unsafe.Pointer(&futureTask.waiters)), uintptr(unsafe.Pointer(node)), uintptr(unsafe.Pointer(nil))) {
			for {
				// signals waiting goroutine
				node.cancel()
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
func (futureTask *FutureTask) awaitDone(timed bool, nanos time.Duration) (int32, error) {
	var deadline = time.Now()
	if timed {
		deadline = deadline.Add(nanos)
	}

	queued := false
	var q *WaitNode = nil
	for {
		if futureTask.runnerCtx.Err() != nil {
			return NIL, InterruptedError
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
		} else if q == nil {
			ctx, cancelFunc := futureTask.createWithCancel()
			q = &WaitNode{gotx: ctx, cancel: cancelFunc}
		} else if !queued {
			q.next = futureTask.waiters
			queued = atomic.CompareAndSwapUintptr((*uintptr)(unsafe.Pointer(&futureTask.waiters)), uintptr(unsafe.Pointer(futureTask.waiters)), uintptr(unsafe.Pointer(q)))
		} else if timed {
			nanos = deadline.Sub(time.Now())
			if nanos <= 0 {
				futureTask.removeWaiter(q)
				return futureTask.state, nil
			}
			select {
			case <-q.gotx.Done():
				break
			case <-time.After(nanos * time.Nanosecond):
				// timeout
			}
		} else {
			// park
			select {
			case <-q.gotx.Done():
				// this node has been unpark
			}
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
				if !atomic.CompareAndSwapUintptr((*uintptr)(unsafe.Pointer(&q)), uintptr(unsafe.Pointer(q)), uintptr(unsafe.Pointer(s))) {
					continue retry
				}
			}
		}
		break
	}
}
