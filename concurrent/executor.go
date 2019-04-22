package concurrent

import (
	"context"
)

type Executor struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewExecutor() *Executor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Executor{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (executor *Executor) Go(executable Executable) Future {
	f := executor.newTaskFor(executable)
	executor.execute(f)
	return f
}

func (executor *Executor) newTaskFor(executable Executable) *FutureTask {
	return NewFutureTask(executor.ctx, executable)
}

func (executor *Executor) execute(f ExecutableFuture) {
	go func() {
		f.Run()
	}()
}

func (executor *Executor) Shutdown() {
	executor.cancel()
}
