package vhost

import (
	"bufio"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/elitah/utils/bufferpool"
)

var (
	ENOCONN = errors.New("no connection can be used")
	ENOBUF  = errors.New("no buffer can be used")

	pool = sync.Pool{
		New: func() interface{} {
			return &sharedConn{}
		},
	}
)

type sharedConn struct {
	sync.Mutex

	net.Conn

	flag uint32

	b *bufferpool.Buffer
}

func (this *sharedConn) Read(p []byte) (n int, err error) {
	if 0x0 != atomic.LoadUint32(&this.flag) {
		return 0, io.EOF
	}

	this.Lock()

	if nil == this.b {
		this.Unlock()

		n, err = this.Conn.Read(p)

		return
	}

	n, err = this.b.Read(p)

	if errors.Is(err, io.EOF) {
		var _n int

		this.b.Free()
		this.b = nil

		_n, err = this.Conn.Read(p[n:])

		n += _n
	}

	this.Unlock()

	return
}

func (this *sharedConn) Write(p []byte) (int, error) {
	if 0x0 != atomic.LoadUint32(&this.flag) {
		return 0, io.ErrClosedPipe
	}
	return this.Conn.Write(p)
}

func (this *sharedConn) Close() error {
	this.Lock()

	if nil != this.b {
		this.b.Free()
		this.b = nil
	}

	this.Unlock()

	if atomic.CompareAndSwapUint32(&this.flag, 0x0, 0x1) {
		if nil != this.Conn {
			this.Conn.Close()
			this.Conn = nil
		}

		pool.Put(this)
	}

	return nil
}

func getConn(conn net.Conn) (_conn *sharedConn, r *bufio.Reader, err error) {
	if b := bufferpool.Get(); nil != b {
		if tr, _err := b.TeeReader(conn, 4*1024); nil == _err {
			if _c := pool.Get(); nil != _c {
				if c, ok := _c.(*sharedConn); ok {
					//
					c.Conn = conn
					c.flag = 0x0
					c.b = b
					//
					return c, bufio.NewReader(tr), nil
				} else {
					err = ENOCONN
				}
			} else {
				err = ENOCONN
			}
		} else {
			err = _err
		}
		b.Free()
	} else {
		err = ENOBUF
	}
	return
}
