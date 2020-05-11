package atomic

import (
	"sync/atomic"
	"unsafe"
)

type AUintptr uintptr

func (this *AUintptr) Add(delta uintptr) uintptr {
	return atomic.AddUintptr((*uintptr)(unsafe.Pointer(this)), delta)
}

func (this *AUintptr) Sub(delta uintptr) uintptr {
	return atomic.AddUintptr((*uintptr)(unsafe.Pointer(this)), 0-delta)
}

func (this *AUintptr) CAS(old, new uintptr) bool {
	return atomic.CompareAndSwapUintptr((*uintptr)(unsafe.Pointer(this)), old, new)
}

func (this *AUintptr) Load() uintptr {
	return atomic.LoadUintptr((*uintptr)(unsafe.Pointer(this)))
}

func (this *AUintptr) Store(val uintptr) {
	atomic.StoreUintptr((*uintptr)(unsafe.Pointer(this)), val)
}

func (this *AUintptr) Swap(new uintptr) uintptr {
	return atomic.SwapUintptr((*uintptr)(unsafe.Pointer(this)), new)
}
