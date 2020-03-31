package httptools

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

var (
	funcs = make(template.FuncMap)

	p1 = &sync.Pool{
		New: func() interface{} {
			return &HTTPWriter{}
		},
	}

	p2 = &sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

type HTTPWriter struct {
	w http.ResponseWriter
	r *http.Request
}

func (this *HTTPWriter) Release() {
	p1.Put(this)
}

func (this *HTTPWriter) GetPath() string {
	return this.r.URL.Path
}

func (this *HTTPWriter) SendHttpRedirect(l string) {
	// 重定向
	this.w.Header().Set("Location", l)
	this.w.WriteHeader(http.StatusFound)
}

func (this *HTTPWriter) SendJSAlert(args ...string) {
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

	fmt.Fprintf(this.w, `<!DOCTYPE html>
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

func (this *HTTPWriter) SendHttpString(s string) {
	// 写HTTP数据
	this.w.Write([]byte(s))
}

func (this *HTTPWriter) SendHttpCode(code int) {
	// 发HTTP状态码
	this.w.WriteHeader(code)
	// 转换HTTP状态码字符串
	if http.StatusOK > code && http.StatusMultipleChoices <= code {
		if s := http.StatusText(code); "" != s {
			this.SendHttpString(s)
		}
	}
}

func (this *HTTPWriter) NotFound() {
	// 发送HTTP状态码：404 Not Found
	this.SendHttpCode(http.StatusNotFound)
}

func (this *HTTPWriter) HttpOnlyIs(methods ...string) bool {
	// 判断是否是指定方法
	for _, method := range methods {
		if method == this.r.Method {
			return true
		}
	}
	// 如果是HEAD方法，如果允许的方法是GET，那么返回200 OK
	if "HEAD" == this.r.Method {
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

func (this *HTTPWriter) TemplateWrite(content []byte, data interface{}, ct string) (bool, error) {
	if b, ok := p2.Get().(*bytes.Buffer); ok {
		// 还
		defer p2.Put(b)
		// 解析模板
		if t, err := template.New(this.GetPath()).Funcs(funcs).Parse(string(content)); nil == err {
			// 复位
			b.Reset()
			// 执行模板
			if err := t.Execute(b, data); nil == err {
				// 发送Content-Type
				if "" != ct {
					this.w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=utf-8", ct))
				}
				// 发送结果
				this.w.Write(b.Bytes())
				// 返回
				return true, nil
			} else {
				return true, err
			}
		} else {
			return true, err
		}
	} else {
		return true, fmt.Errorf("unable get buffer from pool")
	}
}

func (this *HTTPWriter) TemplateFileWrite(path string, data interface{}) (bool, error) {
	// 读取模板文件
	if content, err := ioutil.ReadFile(path); nil == err {
		return this.TemplateWrite(content, data, mime.TypeByExtension(filepath.Ext(path)))
	} else {
		return false, err
	}
}

func NewHTTPWriter(w http.ResponseWriter, r *http.Request) *HTTPWriter {
	if nil != w && nil != r {
		if _r, ok := p1.Get().(*HTTPWriter); ok {
			_r.w = w
			_r.r = r
			return _r
		}
	}
	return nil
}

func TemplateAddFunc(name string, fn interface{}) {
	if "" != name && nil != fn {
		funcs[name] = fn
	}
}
