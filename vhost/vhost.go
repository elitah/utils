package vhost

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
)

var (
	pHTTPConn = sync.Pool{
		New: func() interface{} {
			return &HTTPConn{}
		},
	}

	pReadWriteCloser = sync.Pool{
		New: func() interface{} {
			return &readWriteCloser{}
		},
	}
)

type readWriteCloser struct {
	io.Reader
	io.Writer

	flag uint32

	w io.WriteCloser
}

func (this *readWriteCloser) Read(p []byte) (int, error) {
	if 0x0 != atomic.LoadUint32(&this.flag) {
		return 0, io.ErrClosedPipe
	}
	return this.Reader.Read(p)
}

func (this *readWriteCloser) Write(p []byte) (int, error) {
	if 0x0 != atomic.LoadUint32(&this.flag) {
		return 0, io.ErrClosedPipe
	}
	return this.Writer.Write(p)
}

func (this *readWriteCloser) Close() error {
	if atomic.CompareAndSwapUint32(&this.flag, 0x0, 0x1) {
		if nil != this.w {
			this.w.Close()
			this.w = nil
		}
		pReadWriteCloser.Put(this)
	}
	return nil
}

type HTTPConn struct {
	*sharedConn

	flag uint32

	Method string
	Host   string
	Path   string
}

func (this *HTTPConn) GetTeeReader(w io.WriteCloser) io.ReadWriteCloser {
	if _io, ok := pReadWriteCloser.Get().(*readWriteCloser); ok {
		_io.Reader = io.TeeReader(this, w)
		_io.Writer = this
		_io.flag = 0x0
		_io.w = w
		//
		return _io
	}
	return nil
}

func (this *HTTPConn) Close() error {
	if atomic.CompareAndSwapUint32(&this.flag, 0x0, 0x1) {
		defer pHTTPConn.Put(this)

		return this.sharedConn.Close()
	}
	return nil
}

func HTTP(conn net.Conn) (*HTTPConn, error) {
	if _conn, r, err := getConn(conn); nil == err {
		if nil == r {
			r = bufio.NewReader(_conn)
		}
		if req, err := http.ReadRequest(r); nil == err {
			var method, host, path string
			//
			method = req.Method
			//
			host = req.Host
			//
			if nil != req.URL {
				path = req.URL.Path
			}
			//
			if _c, ok := pHTTPConn.Get().(*HTTPConn); ok {
				_c.sharedConn = _conn
				_c.flag = 0x0
				_c.Method = method
				_c.Host = host
				_c.Path = path
				return _c, nil
			} else {
				return &HTTPConn{
					sharedConn: _conn,
					flag:       0x0,
					Method:     method,
					Host:       host,
					Path:       path,
				}, nil
			}
		} else {
			return nil, fmt.Errorf("http.ReadRequest: %w", err)
		}
	} else {
		return nil, err
	}
}
