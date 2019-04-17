package concurrent

import "errors"

var (
	InterruptedError  = errors.New("goroutine is interrupted")
	CancellationError = errors.New("goroutine is cancelled")
	ExecutionError    = errors.New("goroutine execute error")
	TimeoutError      = errors.New("timeout")
)
