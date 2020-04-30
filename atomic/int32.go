package atomic

import (
	"sync/atomic"
	"unsafe"
)

type AInt32 int32

func (this *AInt32) Add(delta int32) int32 {
	return atomic.AddInt32((*int32)(unsafe.Pointer(this)), delta)
}

func (this *AInt32) CAS(old, new int32) bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(this)), old, new)
}

func (this *AInt32) Load() int32 {
	return atomic.LoadInt32((*int32)(unsafe.Pointer(this)))
}

func (this *AInt32) Store(val int32) {
	atomic.StoreInt32((*int32)(unsafe.Pointer(this)), val)
}

func (this *AInt32) Swap(new int32) int32 {
	return atomic.SwapInt32((*int32)(unsafe.Pointer(this)), new)
}
