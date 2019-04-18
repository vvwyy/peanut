package concurrent

import (
	"time"
)

//type Executable interface {
//	Run() (interface{}, error)
//}

type Executable func () (interface{}, error)

type Future interface {
	Cancel(mayInterruptIfRunning bool) bool
	IsCancelled() bool
	IsDone() bool
	Get() (interface{}, error)
	GetWithTimeout(d time.Duration) (interface{}, error)
}

type ExecutableFuture interface {
	Run()
}

