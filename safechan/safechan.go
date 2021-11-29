package safechan

import (
	"sync/atomic"
)

type ChanCursor struct {
	flag uint32

	max  uint32
	size uint32

	fnClose func()
	fnPush  func(interface{}) bool
}

func NewChanCursor(max, size int, fc func(), fp func(interface{}) bool) *ChanCursor {
	if 0 < max && nil != fc && nil != fp {
		return &ChanCursor{
			max:  uint32(max),
			size: uint32(size),

			fnClose: fc,
			fnPush:  fp,
		}
	}
	//
	return nil
}

func (this *ChanCursor) IsClosed() bool {
	return 0x0 != atomic.LoadUint32(&this.flag)
}

func (this *ChanCursor) Close() {
	if atomic.CompareAndSwapUint32(&this.flag, 0x0, 0x1) {
		this.fnClose()
	}
}

func (this *ChanCursor) Cap() int {
	return int(atomic.LoadUint32(&this.max))
}

func (this *ChanCursor) Len() int {
	return int(atomic.LoadUint32(&this.size))
}

func (this *ChanCursor) Push(v interface{}) bool {
	if nil != v {
		//
		if 0x0 == atomic.LoadUint32(&this.flag) {
			//
			if atomic.LoadUint32(&this.max) >= atomic.AddUint32(&this.size, 1) {
				//
				return this.fnPush(v)
			} else {
				//
				atomic.AddUint32(&this.size, ^uint32(0))
			}
		}
	}
	//
	return false
}

func (this *ChanCursor) Pop() {
	atomic.AddUint32(&this.size, ^uint32(0))
}
