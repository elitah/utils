// +build linux

package random

import (
	"sort"
	"time"
)

func NewRandomTimestamp() int64 {
	var idx int
	var list [32]int

	t := time.Now().UnixNano()

	for ; 32 > idx && 0 < t; t /= 10 {
		//
		list[idx] = int(t % 10)
		//
		if 0 != list[idx] {
			idx++
		}
	}

	sort.Ints(list[:idx])

	t = 0

	for _, item := range list[:idx] {
		if 0 != item {
			t *= 10
			t += int64(item)
		}
	}

	return t
}

func NewRandomTimeNanoSum() int64 {
	var sum int64

	n := time.Now().UnixNano()

	for ; 0 < n; n /= 10 {
		sum += n % 10
	}

	return sum
}
