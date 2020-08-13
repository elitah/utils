package wait

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

type signalOption func(*signalOptions)

type signalOptions struct {
	sigList []os.Signal

	fnNotify func(os.Signal) bool
	fnTicket func(time.Time)

	fnSelect1 func(chan os.Signal)
	fnSelect2 func(chan os.Signal, chan time.Time)
	fnSelect3 func(<-chan os.Signal, <-chan time.Time)

	tInterval int
}

func WithSignal(sigs ...os.Signal) signalOption {
	if 0 < len(sigs) {
		return func(opts *signalOptions) {
			// 标记
			var ok bool
			// 检查
			// 只添加未添加过的信号
			for _, sig := range sigs {
				for _, item := range opts.sigList {
					// 判断信号是否相等
					if ok = signalEqual(item, sig); ok {
						break
					}
				}
				// 将不存在的信号添加
				if !ok {
					opts.sigList = append(opts.sigList, sig)
				}
			}
		}
	}
	return nil
}

func WithNotify(fn func(os.Signal) bool) signalOption {
	if nil != fn {
		return func(opts *signalOptions) {
			opts.fnNotify = fn
		}
	}
	return nil
}

func WithTicket(interval int, fn func(time.Time)) signalOption {
	if 0 < interval {
		return func(opts *signalOptions) {
			opts.fnTicket = fn
			opts.tInterval = interval
		}
	}
	return nil
}

func WithSelect(fn interface{}) signalOption {
	if nil != fn {
		return func(opts *signalOptions) {
			switch result := fn.(type) {
			case func(chan os.Signal):
				opts.fnSelect1 = result
				opts.fnSelect2 = nil
				opts.fnSelect3 = nil
			case func(chan os.Signal, chan time.Time):
				opts.fnSelect1 = nil
				opts.fnSelect2 = result
				opts.fnSelect3 = nil
			case func(<-chan os.Signal, <-chan time.Time):
				opts.fnSelect1 = nil
				opts.fnSelect2 = nil
				opts.fnSelect3 = result
			default:
				panic("WithSelect: received parameter is not valid")
			}
		}
	}
	return nil
}

func Signal(options ...signalOption) (err error) {
	var opts signalOptions

	defer func() {
		if r := recover(); nil != r {
			err = fmt.Errorf("Panic error: %v", r)
		}
	}()

	for _, fn := range options {
		if nil != fn {
			fn(&opts)
		}
	}

	if 0 == len(opts.sigList) {
		opts.sigList = append(opts.sigList, syscall.SIGINT)
	}

	sig := make(chan os.Signal)

	signal.Notify(sig, opts.sigList...)

	defer func() {
		signal.Stop(sig)

		close(sig)
	}()

	if nil != opts.fnSelect1 || nil != opts.fnSelect2 || nil != opts.fnSelect3 {
		_sig := make(chan os.Signal)

		go func() {
			for s := range sig {
				//
				_sig <- s
				//
				if nil != opts.fnNotify {
					if !opts.fnNotify(s) {
						continue
					}
				}
				//
				close(_sig)
				//
				return
			}
		}()

		if nil != opts.fnSelect1 {
			opts.fnSelect1(_sig)
		} else {
			var flag uint32

			_ticker := make(chan time.Time)

			atomic.StoreUint32(&flag, 0x1)

			if 1 <= opts.tInterval {
				ticker := time.NewTicker(time.Duration(opts.tInterval) * time.Second)

				defer ticker.Stop()

				go func() {
					for t := range ticker.C {
						//
						if atomic.CompareAndSwapUint32(&flag, 0x1, 0x2) {
							//
							_ticker <- t
							//
							atomic.StoreUint32(&flag, 0x1)
						}
						//
						if nil != opts.fnTicket {
							opts.fnTicket(t)
						}
					}
				}()
			}

			if nil != opts.fnSelect2 {
				opts.fnSelect2(_sig, _ticker)
			} else {
				opts.fnSelect3(_sig, _ticker)
			}

			for {
				if atomic.CompareAndSwapUint32(&flag, 0x1, 0x0) {
					//
					close(_ticker)
					//
					break
				} else {
					select {
					case <-_ticker:
					default:
					}
				}
			}
		}

		return
	}

	if 1 <= opts.tInterval {
		ticker := time.NewTicker(time.Duration(opts.tInterval) * time.Second)

		defer ticker.Stop()

		for {
			select {
			case s, ok := <-sig:
				//
				if ok {
					if nil != opts.fnNotify {
						if !opts.fnNotify(s) {
							break
						}
					}
				}
				return
			case t, ok := <-ticker.C:
				if ok {
					if nil != opts.fnTicket {
						opts.fnTicket(t)
					}
				}
			}
		}
	} else {
		for s := range sig {
			if nil != opts.fnNotify {
				if !opts.fnNotify(s) {
					continue
				}
			}
			return
		}
	}

	return
}

func signalEqual(s1, s2 os.Signal) bool {
	if _s1, ok := s1.(syscall.Signal); ok {
		if _s2, ok := s2.(syscall.Signal); ok {
			return int(_s1) == int(_s2)
		}
	}
	return false
}
