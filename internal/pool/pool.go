package pool

import (
	"github.com/panjf2000/ants/v2"
)

// Create a new thread/Goroutine pool
func NewThreadPool(numThreads int, opts ...ants.Option) (*ants.Pool, error) {
	return ants.NewPool(numThreads, opts...)
}
