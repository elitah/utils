package captcha

import (
	"bytes"
	"time"

	"github.com/elitah/utils/logs"

	"github.com/dchest/captcha"
)

type CaptchaControl struct {
	mWidth  int
	mHeight int

	mLength int

	nLimit chan byte
}

func NewCaptchaControl(width, height, length int) *CaptchaControl {
	return &CaptchaControl{
		mWidth:  width,
		mHeight: height,

		mLength: length,

		nLimit: make(chan byte, 1),
	}
}

func (this *CaptchaControl) HandleCaptcha(content *bytes.Buffer) (result string) {
	select {
	case this.nLimit <- 1:
		// 生成随机数
		code := captcha.RandomDigits(this.mLength)
		if _, err := captcha.NewImage("", code, this.mWidth, this.mHeight).WriteTo(content); nil == err {
			for i := range code {
				code[i] += '0'
			}
			result = string(code)
		} else {
			logs.Warn("Captcha error: ", err)
		}
		<-this.nLimit
	case <-time.After(3 * time.Second):
	}
	return
}
