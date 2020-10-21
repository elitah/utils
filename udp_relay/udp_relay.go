package udp_relay

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/elitah/utils/atomic"
)

var (
	ENoReady   = errors.New("connection not ready")
	ENoAddress = errors.New("no address could be access")
)

type RelayConn interface {
	net.Conn

	WriteTo([]byte, net.Addr) (int, error)

	String() string
}

type wrapConn struct {
	net.Conn

	remoteAddr net.Addr

	tm_connected atomic.AInt64

	tm_data_send atomic.AInt64
	tm_data_recv atomic.AInt64

	cnt_data_send atomic.AUint64
	cnt_data_recv atomic.AUint64

	cnt_pkg_send atomic.AUint64
	cnt_pkg_recv atomic.AUint64

	cnt_err_send atomic.AUint64
	cnt_err_recv atomic.AUint64
}

func NewWrapConn(conn net.Conn, addrs ...net.Addr) RelayConn {
	//
	wc := &wrapConn{
		Conn: conn,
	}
	//
	if 0 < len(addrs) && nil != addrs[0] {
		wc.remoteAddr = addrs[0]
	}
	//
	wc.tm_connected.Store(time.Now().Unix())
	//
	return wc
}

func (this *wrapConn) Read(data []byte) (n int, err error) {
	//
	n, err = this.Conn.Read(data)
	//
	this.tm_data_recv.Store(time.Now().Unix())
	//
	if 0 < n {
		//
		this.cnt_data_recv.Add(uint64(n))
		//
		this.cnt_pkg_recv.Add(1)
	}
	//
	if nil != err {
		this.cnt_err_recv.Add(1)
	}
	//
	return
}

func (this *wrapConn) Write(data []byte) (n int, err error) {
	//
	this.tm_data_send.Store(time.Now().Unix())
	//
	n, err = this.Conn.Write(data)
	//
	if 0 < n {
		//
		this.cnt_data_send.Add(uint64(n))
		//
		this.cnt_pkg_send.Add(1)
	}
	//
	if nil != err {
		this.cnt_err_send.Add(1)
	}
	//
	return
}

func (this *wrapConn) WriteTo(data []byte, addr net.Addr) (n int, err error) {
	//
	this.tm_data_send.Store(time.Now().Unix())
	//
	if conn, ok := this.Conn.(*net.UDPConn); ok {
		//
		n, err = conn.WriteTo(data, addr)
	} else {
		//
		n, err = this.Conn.Write(data)
	}
	//
	if 0 < n {
		//
		this.cnt_data_send.Add(uint64(n))
		//
		this.cnt_pkg_send.Add(1)
	}
	//
	if nil != err {
		//
		this.cnt_err_send.Add(1)
	}
	//
	return
}

func (this *wrapConn) RemoteAddr() net.Addr {
	//
	if nil != this.remoteAddr {
		//
		return this.remoteAddr
	}
	//
	return this.Conn.RemoteAddr()
}

func (this *wrapConn) SetReadDeadline(t time.Time) error {
	//
	//fmt.Println(time.Now(), time.Since(t))
	//
	return this.Conn.SetReadDeadline(t)
}

func (this *wrapConn) String() string {
	//
	unixnow := time.Now().Unix()
	//
	this.tm_connected.CAS(0, unixnow)
	//
	return fmt.Sprintf(
		"%v: [S]: [A-%v:S-%ds:R-%ds] %d(%d), [R]: %d(%d), [E] %d | %d",
		this.RemoteAddr(),
		time.Duration(unixnow-this.tm_connected.Load()) * time.Second,
		unixnow-this.tm_data_send.Load(),
		unixnow-this.tm_data_recv.Load(),
		this.cnt_data_send.Load(),
		this.cnt_pkg_send.Load(),
		this.cnt_data_recv.Load(),
		this.cnt_pkg_recv.Load(),
		this.cnt_err_send.Load(),
		this.cnt_err_recv.Load(),
	)
	//
	return ""
}

type UDPPacket struct {
	Deadline int64

	Address *net.UDPAddr

	Data   [1024]byte
	Length int

	incoming bool
}

func (this *UDPPacket) Reset() {
	//
	this.Deadline = 0
	this.Address = nil
	this.Length = 0
	//
	this.incoming = false
}

type UDPTransmission struct {
	sync.Mutex

	p *sync.Pool

	m map[string]RelayConn

	c0 chan *UDPPacket
	c1 chan string

	e chan error

	t time.Duration

	f0 func(*net.UDPAddr) RelayConn
	f1 func(*UDPPacket)
	f2 func(error)
}

func NewUDPTransmission(t time.Duration, f0 func(*net.UDPAddr) RelayConn, f1 func(*UDPPacket)) *UDPTransmission {
	if nil != f0 && nil != f1 {
		if p := (&sync.Pool{
			New: func() interface{} {
				return &UDPPacket{}
			},
		}); nil != p {
			//
			if time.Second > t {
				t = time.Second
			}
			//
			//fmt.Println(t)
			//
			u := &UDPTransmission{
				p:  p,
				m:  make(map[string]RelayConn),
				c0: make(chan *UDPPacket, 1024),
				c1: make(chan string),
				e:  make(chan error, 32),
				t:  t,
				f0: f0,
				f1: f1,
			}
			//
			go u.loopSend()
			//
			return u
		}
	}
	return nil
}

func (this *UDPTransmission) SetErrorHandlerFunc(fn func(error)) {
	this.f2 = fn
}

func (this *UDPTransmission) GetUDPPacket() *UDPPacket {
	if p, ok := this.p.Get().(*UDPPacket); ok {
		return p
	}
	return &UDPPacket{}
}

func (this *UDPTransmission) PutUDPPacket(p *UDPPacket) {
	if nil != p {
		//
		p.Reset()
		//
		this.p.Put(p)
	}
}

func (this *UDPTransmission) Forward(p *UDPPacket) bool {
	if nil != p && nil != p.Address {
		//
		p.incoming = false
		//
		this.c0 <- p
		//
		return true
	}
	//
	return false
}

func (this *UDPTransmission) String() string {
	var s strings.Builder
	//
	var list []string
	//
	s.WriteString("--- UDPTransmission ---------------------------------")
	//
	this.Lock()
	//
	if n := len(this.m); 0 < n {
		//
		list = make([]string, n)
		//
		for key, _ := range this.m {
			list = append(list, key)
		}
	}
	//
	this.Unlock()
	//
	if 0 < len(list) {
		//
		sort.Strings(list)
		//
		this.Lock()
		//
		for i, item := range list {
			//
			if conn, ok := this.m[item]; ok {
				if 0 < i {
					s.WriteString("\n")
				}
				fmt.Fprintf(
					&s,
					"%s <===> %v",
					item,
					conn.String(),
				)
			}
		}
		//
		this.Unlock()
	}
	//
	return s.String()
}

func (this *UDPTransmission) loopSend() {
	//
	var conn RelayConn
	//
	for {
		select {
		case p, ok := <-this.c0:
			if ok {
				if !p.incoming {
					//
					token := p.Address.String()
					//
					this.Lock()
					//
					if conn, ok = this.m[token]; ok {
						//
						if _, err := conn.Write(p.Data[:p.Length]); nil == err {
							//
							conn.SetReadDeadline(time.Now().Add(this.t))
						} else {
							//
							ok = false
							//
							this.e <- err
						}
					}
					//
					this.Unlock()
					//
					if !ok {
						//
						go func(token string, p *UDPPacket) {
							//
							if _, err := this.newNode(token, p); nil != err {
								//
								this.e <- err
							}
						}(token, p)
						//
						p = nil
					}
				} else {
					if nil != this.f1 {
						this.f1(p)
					}
				}
				//
				if nil != p {
					this.PutUDPPacket(p)
				}
			}
		case token, ok := <-this.c1:
			if ok {
				//
				this.Lock()
				//
				delete(this.m, token)
				//
				this.Unlock()
			}
		case err, ok := <-this.e:
			if ok {
				if nil != this.f2 {
					this.f2(err)
				}
			}
		}
	}
}

func (this *UDPTransmission) loopRecv(token string, conn net.Conn, address *net.UDPAddr) {
	//
	var err error
	//
	for {
		if p := this.GetUDPPacket(); nil != p {
			//
			conn.SetReadDeadline(time.Now().Add(this.t))
			//
			if p.Length, err = conn.Read(p.Data[:]); nil == err {
				//
				if 0 < p.Length {
					//
					p.Address = address
					//
					p.incoming = true
					//
					this.c0 <- p
					//
					continue
				}
			} else {
				//
				if !errors.Is(err, io.EOF) {
					//
					this.e <- err
				}
				//
				this.PutUDPPacket(p)
				//
				conn.Close()
				//
				break
			}
			//
			this.PutUDPPacket(p)
		}
	}
	//
	this.c1 <- token
}

func (this *UDPTransmission) newNode(token string, p *UDPPacket) (int, error) {
	if "" != token && nil != p {
		//
		if nil != this.f0 {
			//
			if conn := this.f0(p.Address); nil != conn {
				//
				if n, err := conn.Write(p.Data[:p.Length]); nil == err {
					//
					this.Lock()
					//
					this.m[token] = conn
					//
					this.Unlock()
					//
					go this.loopRecv(token, conn, p.Address)
					//
					return n, nil
				} else {
					return 0, err
				}
			} else {
				return 0, ENoReady
			}
		}
	}
	//
	if nil != p {
		//
		this.PutUDPPacket(p)
	}
	//
	return 0, ENoAddress
}
