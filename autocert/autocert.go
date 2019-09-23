package autocert

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

type redirectWriter struct {
	b bytes.Buffer

	header http.Header

	status int
}

func NewRedirectWriter() *redirectWriter {
	return &redirectWriter{
		header: make(http.Header),
	}
}

func (this *redirectWriter) Header() http.Header {
	return this.header
}

func (this *redirectWriter) Write(data []byte) (int, error) {
	return this.b.Write(data)
}

func (this *redirectWriter) WriteHeader(statusCode int) {
	this.status = statusCode
}

func (this *redirectWriter) Redirect(location string) {
	if "" != location {
		this.status = http.StatusFound
		this.header.Set("Location", location)
	}
}

func (this *redirectWriter) Flush(c net.Conn) {
	var length int

	if 0 == this.status {
		this.status = http.StatusOK
	}

	httpStatus := http.StatusText(this.status)

	if "" != httpStatus {
		length = this.b.Len()

		fmt.Fprintf(c, "HTTP/1.1 %d %s\r\n", this.status, httpStatus)
	} else {
		fmt.Fprintf(c, "HTTP/1.1 %d %s\r\n",
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError))
	}

	if 0 < length {
		this.header.Set("Content-Length", fmt.Sprint(length))

		if contentType := this.header.Get("Content-Type"); "" == contentType {
			this.header.Set("Content-Type", "text/plain; charset=utf-8")
		}
	} else {
		this.header.Del("Content-Length")
		this.header.Del("Content-Type")
	}

	this.header.Set("Date", time.Now().Format(time.RFC1123))

	this.header.Write(c)

	fmt.Fprint(c, "\r\n")

	if 0 < length {
		io.Copy(c, &this.b)
	}
}

type AutoCertManager struct {
	*autocert.Manager

	handler http.Handler
}

func NewAutoCertManager() *AutoCertManager {
	var cacheDir string

	if file, err := exec.LookPath(os.Args[0]); nil == err {
		if path, err := filepath.Abs(file); nil == err {
			if cacheDir = filepath.Dir(path); "" != cacheDir {
				cacheDir += "/letscache"
			}
		}
	}

	if "" != cacheDir {
		_acm := &autocert.Manager{
			Cache:    autocert.DirCache(cacheDir),
			Prompt:   autocert.AcceptTOS,
			ForceRSA: true,
		}

		acm := &AutoCertManager{
			Manager: _acm,
		}

		if acm.handler = _acm.HTTPHandler(acm); nil != acm.handler {
			return acm
		}
	}

	return nil
}

func (this *AutoCertManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 强制调整到HTTPS
	if r.Method == "GET" || r.Method == "HEAD" {
		if _w, ok := w.(*redirectWriter); ok {
			if host, _, err := net.SplitHostPort(r.Host); nil != err {
				_w.Redirect("https://" + r.Host + r.URL.RequestURI())
			} else {
				_w.Redirect("https://" + host + r.URL.RequestURI())
			}
			return
		}
	}
	// 输出错误，Bad Request
	w.WriteHeader(http.StatusBadRequest)
}

func (this *AutoCertManager) ServeRequest(w net.Conn, r *http.Request) {
	_w := NewRedirectWriter()

	if nil != this.handler {
		this.handler.ServeHTTP(_w, r)
	} else {
		this.ServeHTTP(_w, r)
	}

	_w.Flush(w)
}
