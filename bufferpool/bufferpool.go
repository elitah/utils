package bufferpool

import (
	"bytes"
	"errors"
	"io"
	"sync"
)

var (
	ENOBUF = errors.New("no such buffer can be used")

	GlobalPool = NewBufferPool(32)

	pReadCloser = sync.Pool{
		New: func() interface{} {
			return &readCloser{}
		},
	}

	pLimitedReader = sync.Pool{
		New: func() interface{} {
			return &io.LimitedReader{}
		},
	}

	pLimitedWriter = sync.Pool{
		New: func() interface{} {
			return &limitedWriter{}
		},
	}
)

type BufferPool struct {
	*sync.Pool

	blackSize int
}

func (this *BufferPool) New() interface{} {
	if data := make([]byte, 0, this.blackSize*1024); nil != data {
		if b := bytes.NewBuffer(data); nil != b {
			return &Buffer{
				Buffer: b,
				pool:   this,
			}
		}
	}
	return nil
}

func (this *BufferPool) Get() *Buffer {
	if _b := this.Pool.Get(); nil != _b {
		if b, ok := _b.(*Buffer); ok {
			//
			b.AddRefer(1)
			//
			return b
		}
	}
	return nil
}

func (this *BufferPool) Put(b *Buffer) {
	b.Free()
}

type limitedWriter struct {
	w   io.Writer
	max int64
}

func (this *limitedWriter) Write(p []byte) (int, error) {
	if f, ok := this.w.(interface {
		Len() int
	}); ok {
		if n := int64(f.Len()); this.max > n {
			// 计算数据长度
			if _n := int64(len(p)); this.max-n > _n {
				n = _n
			} else {
				n = this.max - n
			}
			return this.w.Write(p[:int(n)])
		}
	}
	return len(p), nil
}

type readCloser struct {
	io.Reader

	lw *limitedWriter
}

func (this *readCloser) Close() error {
	//
	if nil != this.lw {
		pLimitedWriter.Put(this.lw)
		this.lw = nil
	}
	//
	pReadCloser.Put(this)
	//
	return nil
}

type Buffer struct {
	*bytes.Buffer

	refcnt SCounter
	usecnt SCounter

	pool *BufferPool
}

func (this *Buffer) ReadFromLimited(r io.Reader, n int64) (int64, error) {
	if lr, ok := pLimitedReader.Get().(*io.LimitedReader); ok {
		//
		defer pLimitedReader.Put(lr)
		//
		lr.R = r
		lr.N = n
		//
		return this.ReadFrom(lr)
	}
	return 0, ENOBUF
}

func (this *Buffer) TeeReader(r io.Reader, max int64) (io.ReadCloser, error) {
	if _r, ok := pReadCloser.Get().(*readCloser); ok {
		if lw, ok := pLimitedWriter.Get().(*limitedWriter); ok {
			//
			lw.w = this
			lw.max = max
			//
			_r.Reader = io.TeeReader(r, lw)
			_r.lw = lw
			//
			return _r, nil
		}
	}
	return nil, ENOBUF
}

func (this *Buffer) IsFree() bool {
	return 0 >= this.refcnt.Load()
}

func (this *Buffer) AddRefer(n int) int64 {
	this.usecnt.Add(n)
	return int64(this.refcnt.Add(n))
}

func (this *Buffer) Free() (n int64) {
	if n = int64(this.refcnt.Add(-1)); 0 == n {
		this.Reset()
		this.pool.Pool.Put(this)
	}
	return
}

func NewBufferPool(blackSize int) *BufferPool {
	p := &BufferPool{
		blackSize: blackSize,
	}

	p.Pool = &sync.Pool{
		New: p.New,
	}

	return p
}

func Get() *Buffer {
	return GlobalPool.Get()
}

func Put(b *Buffer) {
	GlobalPool.Put(b)
}
