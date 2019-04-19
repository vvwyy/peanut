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
	//defer func() {
	//	if r := recover(); r != nil {
	//
	//		fmt.Println("Recovered in f", r)
	//	}
	//}()

	//f, err := executor.newTaskFor(executable)
	//if err != nil {
	//	return nil, err
	//}
	//err = executor.execute(f)
	//if err != nil {
	//	return nil, err
	//}
	//return f, nil


	f := executor.newTaskFor(executable)
	executor.execute(f)
	return f

}

func (executor *Executor) newTaskFor(executable Executable) *FutureTask {
	//if executable == nil {
	//	return nil, errors.New("executable is nil")
	//}
	return NewFutureTask(executor.ctx, executable)
}

func (executor *Executor) execute(f ExecutableFuture) {
	//if f == nil {
	//	return errors.New("executable future is nil")
	//}
	go func() {
		f.Run()
	}()

	//return nil
}

func (executor *Executor) Shutdown() {
	executor.cancel()
}
