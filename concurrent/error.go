package concurrent

import "github.com/pkg/errors"

var (
	InterruptedError = errors.New("Goroutine is interrupted.")
	CancellationError = errors.New("Goroutine is cancelled.")
	ExecutionError = errors.New("Goroutine execute error")
)
