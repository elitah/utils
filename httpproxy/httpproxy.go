package httpproxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/elitah/fast-io"
	"github.com/elitah/utils/logs"
)

type HttpProxy struct {
	ConnectTimeout time.Duration

	Hijack func(net.Conn, *http.Request, string, string) bool

	DialContext func(context.Context, string, string) (net.Conn, error)
}

func (this *HttpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if "" != r.URL.Host {
		// 判断是否运行劫持
		if hj, ok := w.(http.Hijacker); ok {
			// 得到原始连接
			if conn_remote, _, err := hj.Hijack(); nil == err {
				// 退出时关闭原始连接
				defer conn_remote.Close()
				// 得到域名和端口
				host, port, _ := net.SplitHostPort(r.URL.Host)
				// 域名为空，说明地址不包含端口
				if "" == host {
					host = r.URL.Host
				}
				// 当端口为空时进行模式设定
				if "" == port {
					if "https" == r.URL.Scheme {
						port = "443"
					} else {
						port = "80"
					}
				}
				// 劫持
				if nil != this.Hijack {
					if this.Hijack(conn_remote, r, host, port) {
						return
					}
				}
				// 建立连接
				if conn_local, err := this.Dial(r.Context(), "tcp", host, port); nil == err {
					// 退出时关闭代理连接
					defer conn_local.Close()
					// 判断请求模式
					if "CONNECT" == r.Method {
						conn_remote.Write([]byte("HTTP/1.0 200 Connection Established\r\n\r\n"))
					} else {
						r.URL.Scheme = ""
						r.URL.Opaque = ""
						//r.URL.User = nil
						r.URL.Host = ""

						r.Write(conn_local)
					}

					fast_io.FastCopy(conn_remote, conn_local)

					return
				} else {
					logs.Error(err)
				}
			} else {
				logs.Error(err)
			}
		} else {
			logs.Error("unable to hijack connection")
		}
	} else {
		logs.Error("unable get target address")
	}
}

func (this *HttpProxy) Dial(ctx context.Context, network, host, port string) (net.Conn, error) {
	// 修正代理接口
	if nil == this.DialContext {
		this.DialContext = (&net.Dialer{}).DialContext
	}

	if time.Second <= this.ConnectTimeout {
		// 检查context是否存在有效期
		// 如果不存在则指定超时时间
		if _, ok := ctx.Deadline(); !ok {
			//
			var cancel context.CancelFunc
			//
			ctx, cancel = context.WithTimeout(ctx, this.ConnectTimeout)
			//
			defer cancel()
		}
	}

	return this.DialContext(ctx, network, fmt.Sprintf("%s:%s", host, port))
}
