package errors

import (
	"errors"
	"fmt"
	"io"
)

func IsEOF(err error) bool {
	return errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF)
}

func IsTimeout(err error) bool {
	for {
		//
		if x, ok := err.(interface {
			IsTimeout() bool
		}); ok {
			return x.IsTimeout()
		}
		//
		if x, ok := err.(interface {
			Timeout() bool
		}); ok {
			return x.Timeout()
		}
		//
		if err = errors.Unwrap(err); err == nil {
			return false
		}
	}
}

func TryCatchPanic(fn func()) (err error) {
	//
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Panic error: %v", r)
		}
	}()
	//
	fn()
	//
	return
}
