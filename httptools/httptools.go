package httptools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/elitah/fast-io"
)

const (
	flagDebug = iota

	flagMax

	flagDebugEnabled
	flagDebugDisabled
)

var (
	funcs = make(template.FuncMap)

	p1 = &sync.Pool{
		New: func() interface{} {
			return &httpHandler{
				Request: nil,
			}
		},
	}

	p2 = &sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

type httpHandler struct {
	*http.Request

	flags [flagMax]uint32

	statusCode int

	location    string
	contentType string

	start int64

	rb *bytes.Buffer
	wb *bytes.Buffer

	jenc *json.Encoder
}

func (this *httpHandler) Release() {
	this.Body.Close()

	p2.Put(this.rb)
	p2.Put(this.wb)

	this.rb = nil
	this.wb = nil

	this.jenc = nil

	p1.Put(this)
}

func (this *httpHandler) Debug(flag bool) {
	if flag {
		atomic.StoreUint32(&this.flags[flagDebug], flagDebugEnabled)
		return
	}
	atomic.StoreUint32(&this.flags[flagDebug], flagDebugDisabled)
}

func (this *httpHandler) GetPath() string {
	return this.URL.Path
}

func (this *httpHandler) GetJson(v interface{}) error {
	if 0 < this.rb.Len() {
		//
		return json.Unmarshal(this.rb.Bytes(), v)
	}
	return nil
}

func (this *httpHandler) SendHttpCode(code int) {
	this.statusCode = code
}

func (this *httpHandler) SendHttpRedirect(l string) {
	// 重定向
	this.location = l
}

func (this *httpHandler) NotFound() {
	// 发送HTTP状态码：404 Not Found
	this.SendHttpCode(http.StatusNotFound)
}

func (this *httpHandler) HttpOnlyIs(methods ...string) bool {
	// 判断是否是指定方法
	for _, method := range methods {
		if method == this.Method {
			return true
		}
	}
	// 如果是HEAD方法，如果允许的方法是GET，那么返回200 OK
	if "HEAD" == this.Method {
		for _, method := range methods {
			if "GET" == method {
				this.SendHttpCode(http.StatusOK)
				return false
			}
		}
	}
	// 发送HTTP状态码：405 Method Not Allowed
	this.SendHttpCode(http.StatusMethodNotAllowed)
	// 返回false
	return false
}

func (this *httpHandler) SendJSAlert(args ...string) {
	var title, msg, redirect string = "提示", "未填写消息内容", "/"

	if 1 <= len(args) && "" != args[0] {
		title = args[0]
	}

	if 2 <= len(args) && "" != args[1] {
		msg = args[1]
	}

	if 3 <= len(args) && "" != args[2] {
		redirect = args[2]
	}

	// 清空缓冲
	this.wb.Reset()

	fmt.Fprintf(this.wb, `<!DOCTYPE html>
<html lang="zh-cn">
	<head>
		<title>%s</title>
	</head>
	<body>
		<script>
		alert('%s');
		window.location.href = '%s';
		</script>
	</body>
</html>
`, title, msg, redirect)
}

func (this *httpHandler) SendHttpString(s string) {
	// 清空缓冲
	this.wb.Reset()
	// 写HTTP数据
	this.wb.WriteString(s)
}

func (this *httpHandler) SendJsonString(s string) {
	// 清空缓冲
	this.wb.Reset()
	// 设置ContentType
	this.contentType = "application/json"
	// 写Json数据
	this.wb.WriteString(s)
}

func (this *httpHandler) SendJson(v interface{}) error {
	if nil == this.jenc {
		this.jenc = json.NewEncoder(this.wb)
	}
	if nil != this.jenc {
		// 清空缓冲
		this.wb.Reset()
		// 设置ContentType
		this.contentType = "application/json"
		// 输出
		return this.jenc.Encode(v)
	}
	// 输出json
	if data, err := json.Marshal(v); nil == err {
		if _, err = this.wb.Write(data); nil == err {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

func (this *httpHandler) TemplateWrite(content []byte, data interface{}, ct string) error {
	// 解析模板
	if t, err := template.New(this.GetPath()).Funcs(funcs).Parse(string(content)); nil == err {
		// 复位
		this.wb.Reset()
		// 执行模板
		if err := t.Execute(this.wb, data); nil == err {
			// 设置Content-Type
			this.contentType = ct
			// 返回
			return nil
		} else {
			//
			this.statusCode = http.StatusInternalServerError
			// 模板错误，复位
			this.wb.Reset()
			// 返回错误
			return err
		}
	} else {
		// 返回错误
		return err
	}
}

func (this *httpHandler) TemplateFileWrite(path string, data interface{}) (bool, error) {
	// 打开模板文件
	if f, err := os.Open(path); nil == err {
		// 退出是关闭文件
		defer f.Close()
		// 得到buffer
		if b, ok := p2.Get().(*bytes.Buffer); ok {
			// 退出是释放buffer
			defer p2.Put(b)
			// 快速拷贝
			if _, err = fast_io.Copy(b, f); nil == err {
				// 加载模板数据
				return true, this.TemplateWrite(b.Bytes(), data, mime.TypeByExtension(filepath.Ext(path)))
			} else {
				return true, err
			}
		} else {
			return true, fmt.Errorf("no buffer can be used")
		}
	} else {
		return false, err
	}
}

func (this *httpHandler) Output(w http.ResponseWriter) string {
	var debug *bytes.Buffer

	if flagDebugEnabled == atomic.LoadUint32(&this.flags[flagDebug]) {
		if b, ok := p2.Get().(*bytes.Buffer); ok {
			//
			defer p2.Put(b)
			//
			b.Reset()
			//
			b.WriteString("================================================\n")
			//
			fmt.Fprintf(b, "Action: %s | %s(%s) | %s | %s\n", this.Method, this.Host, this.URL.Host, this.GetPath(), this.URL.RequestURI())
			fmt.Fprintf(b, "Cost: %.3f ms\n", float64((time.Now().UnixNano()/1000)-this.start)/1000.0)
			fmt.Fprintf(b, "Flags: %v\n", this.flags)
			fmt.Fprintf(b, "StatusCode: %d\n", this.statusCode)
			fmt.Fprintf(b, "Location: %s\n", this.location)
			fmt.Fprintf(b, "ContentType: %s\n", this.contentType)
			fmt.Fprintf(b, "User-Agent: %s\n", this.UserAgent())
			fmt.Fprintf(b, "rb: %s\n", this.rb.String())
			fmt.Fprintf(b, "wb: %s\n", this.wb.String())
			//
			b.WriteString("\n")
			//
			debug = b
		}
	}

	defer func() {
		// 发HTTP状态码
		w.WriteHeader(this.statusCode)
		// 写数据
		if 0 < this.wb.Len() {
			w.Write(this.wb.Bytes())
		}
	}()

	if "" != this.location {
		// 发HTTP状态码
		this.statusCode = http.StatusFound
		// 写location跳转路径
		w.Header().Set("Location", this.location)
	} else {
		// 写Content-Type
		if "" != this.contentType {
			if "text/json" != this.contentType && strings.Contains(this.contentType, "/json") {
				if strings.Contains(this.UserAgent(), "MSIE") {
					this.contentType = "text/json"
				}
			}
			//
			switch this.contentType {
			case "text/html", "text/css", "text/javascript", "application/x-javascript", "text/json", "application/json":
				this.contentType = fmt.Sprintf("%s; charset=utf-8", this.contentType)
			default:
			}
			//
			w.Header().Set("Content-Type", this.contentType)
		} else {
			//
			if 0 < this.wb.Len() {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
			} else {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			}
		}
		// 写缓冲
		if http.StatusBadRequest <= this.statusCode {
			if s := http.StatusText(this.statusCode); "" != s {
				// 缓冲复位
				this.wb.Reset()
				// 重写缓冲
				this.wb.WriteString(s)
			}
		}
	}
	//
	if nil != debug && 0 < debug.Len() {
		return debug.String()
	}
	//
	return ""
}

func NewHttpHandler(r *http.Request) *httpHandler {
	if nil != r {
		if rb, ok := p2.Get().(*bytes.Buffer); ok {
			if wb, ok := p2.Get().(*bytes.Buffer); ok {
				if _r, ok := p1.Get().(*httpHandler); ok {
					//
					rb.Reset()
					wb.Reset()
					//
					if "POST" == r.Method {
						//
						var buffer [1024]byte
						//
						r.ParseForm()
						//
						io.CopyBuffer(rb, r.Body, buffer[:])
					}
					//
					_r.Request = r
					//
					atomic.StoreUint32(&_r.flags[flagDebug], flagDebugDisabled)
					//
					_r.statusCode = http.StatusOK
					_r.location = ""
					_r.contentType = ""
					_r.start = time.Now().UnixNano() / 1000
					// 缓冲区
					_r.rb = rb
					_r.wb = wb
					// json编码器
					_r.jenc = nil
					// 返回对象
					return _r
				}
				// 释放
				p2.Put(wb)
			}
			// 释放
			p2.Put(rb)
		}
	}
	return nil
}

func TemplateAddFunc(name string, fn interface{}) {
	if "" != name && nil != fn {
		funcs[name] = fn
	}
}
