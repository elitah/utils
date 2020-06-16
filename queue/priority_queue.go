package queue

import (
	"container/heap"
	"sync"
	"sync/atomic"
)

type QItem interface {
	GetPriority() int64
}

type PriorityQueue struct {
	sync.Mutex

	flag uint32

	list []QItem
}

func (this *PriorityQueue) Init() {
	if atomic.CompareAndSwapUint32(&this.flag, 0x0, 0x1) {
		this.Lock()
		defer this.Unlock()

		heap.Init(this)
	}
}

func (this *PriorityQueue) Len() int {
	return len(this.list)
}

func (this *PriorityQueue) Cap() int {
	return cap(this.list)
}

func (this *PriorityQueue) Less(i, j int) bool {
	return this.list[i].GetPriority() < this.list[j].GetPriority()
}

func (this *PriorityQueue) Swap(i, j int) {
	this.list[i], this.list[j] = this.list[j], this.list[i]
}

func (this *PriorityQueue) Push(x interface{}) {
	if item, ok := x.(QItem); ok {
		this.list = append(this.list, item)
	}
}

func (this *PriorityQueue) Pop() (result interface{}) {
	n := len(this.list)
	result = this.list[n-1]
	this.list[n-1] = nil
	this.list = this.list[0 : n-1]
	return
}

func (this *PriorityQueue) PushItem(item QItem) {
	this.Lock()
	defer this.Unlock()

	heap.Push(this, item)
}

func (this *PriorityQueue) PopItem() QItem {
	this.Lock()
	defer this.Unlock()

	for 0 < this.Len() {
		if item, ok := heap.Pop(this).(QItem); ok {
			return item
		}
	}
	return nil
}
