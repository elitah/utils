// +build arm

package bufferpool

import (
	"sync/atomic"
)

type SCounter struct {
	v int32
}

func (this *SCounter) Add(n int) int32 {
	return atomic.AddInt32(&this.v, int32(n))
}

func (this *SCounter) Load() int32 {
	return atomic.LoadInt32(&this.v)
}
