package atomic

import (
	"sync/atomic"
	"unsafe"
)

type AUint32 uint32

func (this *AUint32) Add(delta uint32) uint32 {
	return atomic.AddUint32((*uint32)(unsafe.Pointer(this)), delta)
}

func (this *AUint32) CAS(old, new uint32) bool {
	return atomic.CompareAndSwapUint32((*uint32)(unsafe.Pointer(this)), old, new)
}

func (this *AUint32) Load() uint32 {
	return atomic.LoadUint32((*uint32)(unsafe.Pointer(this)))
}

func (this *AUint32) Store(val uint32) {
	atomic.StoreUint32((*uint32)(unsafe.Pointer(this)), val)
}

func (this *AUint32) Swap(new uint32) uint32 {
	return atomic.SwapUint32((*uint32)(unsafe.Pointer(this)), new)
}
