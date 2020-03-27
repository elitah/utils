package httptools

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"text/template"
)

type HTTPWriter struct {
	w http.ResponseWriter
	r *http.Request
}

func NewHTTPWriter(w http.ResponseWriter, r *http.Request) *HTTPWriter {
	if nil != w && nil != r {
		return &HTTPWriter{w, r}
	}
	return nil
}

func (this *HTTPWriter) GetPath() string {
	return this.r.URL.Path
}

func (this *HTTPWriter) SendHttpRedirect(l string) {
	// 重定向
	this.w.Header().Set("Location", l)
	this.w.WriteHeader(http.StatusFound)
}

func (this *HTTPWriter) SendHttpString(s string) {
	// 写HTTP数据
	this.w.Write([]byte(s))
}

func (this *HTTPWriter) SendHttpCode(code int) {
	// 转换HTTP状态码字符串
	if s := http.StatusText(code); "" != s {
		this.w.WriteHeader(code)
		this.SendHttpString(s)
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

func (this *HTTPWriter) TemplateWrite(w http.ResponseWriter, content []byte, data interface{}, ct string) (bool, error) {
	// 解析模板
	if t, err := template.New(this.GetPath()).Parse(string(content)); nil == err {
		// 缓冲
		var b bytes.Buffer
		// 执行模板
		if err := t.Execute(&b, data); nil == err {
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
}

func (this *HTTPWriter) TemplateFileWrite(w http.ResponseWriter, path string, data interface{}) (bool, error) {
	// 读取模板文件
	if content, err := ioutil.ReadFile(path); nil == err {
		return this.TemplateWrite(w, content, data, mime.TypeByExtension(filepath.Ext(path)))
	} else {
		return false, err
	}
}
