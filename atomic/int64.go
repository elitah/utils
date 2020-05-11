package atomic

import (
	"sync/atomic"
	"unsafe"
)

type AInt64 int64

func (this *AInt64) Add(delta int64) int64 {
	return atomic.AddInt64((*int64)(unsafe.Pointer(this)), delta)
}

func (this *AInt64) Sub(delta int64) int64 {
	return atomic.AddInt64((*int64)(unsafe.Pointer(this)), 0-delta)
}

func (this *AInt64) CAS(old, new int64) bool {
	return atomic.CompareAndSwapInt64((*int64)(unsafe.Pointer(this)), old, new)
}

func (this *AInt64) Load() int64 {
	return atomic.LoadInt64((*int64)(unsafe.Pointer(this)))
}

func (this *AInt64) Store(val int64) {
	atomic.StoreInt64((*int64)(unsafe.Pointer(this)), val)
}

func (this *AInt64) Swap(new int64) int64 {
	return atomic.SwapInt64((*int64)(unsafe.Pointer(this)), new)
}
