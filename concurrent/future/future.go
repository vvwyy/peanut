package future

import (
	"time"
)

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

