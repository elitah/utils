// +build linux

package random

import (
	"errors"
	"io/ioutil"
)

var (
	ENoUUID = errors.New("Unable read uuid from kernel")
)

func NewRandomUUIDByKernel() string {
	if content, err := ioutil.ReadFile("/proc/sys/kernel/random/uuid"); nil == err {
		if 36 <= len(content) {
			return string(content[:36])
		}
	}
	panic(ENoUUID)
}
