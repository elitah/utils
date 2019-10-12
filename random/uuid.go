// +build linux

package random

import (
	"errors"
	"fmt"
	"io/ioutil"
	"runtime"
)

var (
	ENoUUID       = errors.New("Unable read uuid")
	ENoUUIDKernel = errors.New("Unable read uuid from kernel")
)

func NewRandomUUID() string {
	s1 := NewRandomString(ModeHexLower, 8)
	s2 := NewRandomString(ModeHexLower, 4)
	s3 := NewRandomString(ModeHexLower, 4)
	s4 := NewRandomString(ModeHexLower, 4)
	s5 := NewRandomString(ModeHexLower, 12)
	if "" != s1 && "" != s2 && "" != s3 && "" != s4 && "" != s5 {
		return fmt.Sprintf("%s-%s-%s-%s-%s", s1, s2, s3, s4, s5)
	}
	panic(ENoUUID)
}

func NewRandomUUIDByKernel() string {
	if "linux" == runtime.GOOS {
		if content, err := ioutil.ReadFile("/proc/sys/kernel/random/uuid"); nil == err {
			if 36 <= len(content) {
				return string(content[:36])
			}
		}
		panic(ENoUUIDKernel)
	}
	return NewRandomUUID()
}
