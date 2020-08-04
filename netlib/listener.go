package netlib

import (
	"errors"
	"net"

	"github.com/elitah/utils/atomic"
)

var (
	EClosed = errors.New("channel closed")
)

type ListenerWithInput interface {
	net.Listener

	Input(net.Conn) error
}

type chanListener struct {
	flag atomic.AInt32

	addr net.Addr

	ch chan net.Conn
}

func NewChanListener(addr net.Addr, size int) ListenerWithInput {
	if nil != addr {
		return &chanListener{
			addr: addr,
			ch:   make(chan net.Conn, size),
		}
	}
	return nil
}

func (this *chanListener) Accept() (net.Conn, error) {
	if 0x0 == this.flag.Load() {
		if conn, ok := <-this.ch; ok {
			return conn, nil
		}
	}
	return nil, EClosed
}

func (this *chanListener) Addr() net.Addr {
	return this.addr
}

func (this *chanListener) Input(conn net.Conn) error {
	if 0x0 == this.flag.Load() {
		this.ch <- conn
		return nil
	}
	return EClosed
}

func (this *chanListener) Close() error {
	if this.flag.CAS(0x0, 0x1) {
		//
		close(this.ch)
		//
		this.ch = nil
		//
		return nil
	}
	//
	return EClosed
}
