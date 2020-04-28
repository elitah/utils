package wait

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

type signalOption func(*signalOptions)

type signalOptions struct {
	sigList []os.Signal

	fnNotify func(os.Signal) bool
	fnTicket func(time.Time)

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
	if nil != fn {
		return func(opts *signalOptions) {
			opts.fnTicket = fn

			if 1 <= interval {
				opts.tInterval = interval
			} else {
				opts.tInterval = 3
			}
		}
	}
	return nil
}

func Signal(options ...signalOption) {
	var opts signalOptions

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

	if nil != opts.fnTicket && 1 <= opts.tInterval {
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
					opts.fnTicket(t)
				}
			}
		}
	} else {
		for {
			if s, ok := <-sig; ok {
				if nil != opts.fnNotify {
					if !opts.fnNotify(s) {
						continue
					}
				}
			}
			return
		}
	}
}

func signalEqual(s1, s2 os.Signal) bool {
	if _s1, ok := s1.(syscall.Signal); ok {
		if _s2, ok := s2.(syscall.Signal); ok {
			return int(_s1) == int(_s2)
		}
	}
	return false
}
