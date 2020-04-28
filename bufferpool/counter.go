// +build !arm

package bufferpool

import (
	"sync/atomic"
)

type SCounter struct {
	v int64
}

func (this *SCounter) Add(n int) int64 {
	return atomic.AddInt64(&this.v, int64(n))
}

func (this *SCounter) Load() int64 {
	return atomic.LoadInt64(&this.v)
}
