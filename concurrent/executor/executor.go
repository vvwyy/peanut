package executor

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
)

type Executor struct {
	ctx                   context.Context
	cancel                context.CancelFunc
	activeGoroutinesMutex *sync.Mutex
	activeGoroutines      map[string]int
	HandlePanic           func(recovered interface{}, funcName string)
}

type Executable interface {
	//Group() string
	Run() (interface{}, error)
}

func NewExecutor() *Executor {
	ctx, cancel := context.WithCancel(context.TODO())
	return &Executor{
		ctx:                   ctx,
		cancel:                cancel,
		activeGoroutinesMutex: &sync.Mutex{},
		activeGoroutines:      map[string]int{},
	}
}

//func (executor *Executor) Go(handler func(ctx context.Context) interface{}) {
//	executor.activeGoroutinesMutex.Lock()
//	defer executor.activeGoroutinesMutex.Unlock()
//}

func (executor *Executor) Go(executable Executable) (future *Future) {
	future = newFuture(executable)
	executor.execute(future)
	return
}

func (executor *Executor) execute(future *Future) {
	executor.activeGoroutinesMutex.Lock()
	defer executor.activeGoroutinesMutex.Unlock()
	executor.activeGoroutines[executable.Group()] += 1
	var (
		result interface{}
		err    error
	)
	go func() {
		defer func() {
			recovered := recover()
			if recovered != nil {
				if executor.HandlePanic == nil {
					handlePanic(recovered, executable.Group())
				} else {
					executor.HandlePanic(recovered, executable.Group())
				}
			}
			executor.activeGoroutinesMutex.Lock()
			executor.activeGoroutines[executable.Group()] -= 1
			executor.activeGoroutinesMutex.Unlock()
		}()
		result, err = executable.Run()
	}()
	future <- &Future{
		parent: executor.ctx,
		result: result,
		err:    err,
	}
	return
}






func (executor *Executor) Stop() {
	executor.cancel()
}

func handlePanic(recovered interface{}, funcName string) {
	log.Println(fmt.Sprintf("%s panic: %v", funcName, recovered))
	log.Println(string(debug.Stack()))
}
