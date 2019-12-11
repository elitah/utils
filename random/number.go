package random

import (
	"crypto/rand"
	"math/big"
)

func NewRandomInt32(max int32) int32 {
	if r, err := rand.Int(rand.Reader, big.NewInt(int64(max))); nil == err {
		return int32(r.Int64())
	}
	return 0
}

func NewRandomUint32(max uint32) uint32 {
	if r, err := rand.Int(rand.Reader, big.NewInt(int64(max))); nil == err {
		return uint32(r.Uint64())
	}
	return 0
}

func NewRandomInt64(max int64) int64 {
	if r, err := rand.Int(rand.Reader, big.NewInt(int64(max))); nil == err {
		return r.Int64()
	}
	return 0
}

func NewRandomUint64(max uint64) uint64 {
	if r, err := rand.Int(rand.Reader, big.NewInt(int64(max))); nil == err {
		return r.Uint64()
	}
	return 0
}

func NewRandomInt(max int) int {
	return int(NewRandomInt32(int32(max)))
}

func NewRandomUint(max uint) uint {
	return uint(NewRandomUint32(uint32(max)))
}
