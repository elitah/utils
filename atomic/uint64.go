package atomic

import (
	"sync/atomic"
	"unsafe"
)

type AUint64 uint64

func (this *AUint64) Add(delta uint64) uint64 {
	return atomic.AddUint64((*uint64)(unsafe.Pointer(this)), delta)
}

func (this *AUint64) Sub(delta uint64) uint64 {
	return atomic.AddUint64((*uint64)(unsafe.Pointer(this)), ^uint64(delta-1))
}

func (this *AUint64) CAS(old, new uint64) bool {
	return atomic.CompareAndSwapUint64((*uint64)(unsafe.Pointer(this)), old, new)
}

func (this *AUint64) Load() uint64 {
	return atomic.LoadUint64((*uint64)(unsafe.Pointer(this)))
}

func (this *AUint64) Store(val uint64) {
	atomic.StoreUint64((*uint64)(unsafe.Pointer(this)), val)
}

func (this *AUint64) Swap(new uint64) uint64 {
	return atomic.SwapUint64((*uint64)(unsafe.Pointer(this)), new)
}
