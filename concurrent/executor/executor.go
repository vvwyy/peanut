package executor

import (
	"context"
	"errors"
	"github.com/vvwyy/peanut/concurrent/future"
)

type Executor struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type Executable interface {
	Run() (interface{}, error)
}

func NewExecutor() *Executor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Executor{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (executor *Executor) Go(executable Executable) (future.Future, error) {
	f, err := executor.newTaskFor(executable)
	if err != nil {
		return nil, err
	}
	err = executor.execute(f)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (executor *Executor) newTaskFor(executable Executable) (*future.FutureTask, error) {
	if executable == nil {
		return nil, errors.New("executable is nil")
	}
	return future.NewFutureTask(executable), nil
}

func (executor *Executor) execute(f future.ExecutableFuture) error {
	if f == nil {
		return errors.New("executable future is nil")
	}
	go func() {
		f.Run(executor.ctx)
	}()

	return nil
}

func (executor *Executor) Shutdown() {
	executor.cancel()
}

/*
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



func (executor *Executor) execute(future future.ExecutableFuture) {
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
*/
