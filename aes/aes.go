package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
)

type AESTool struct {
	bytes.Buffer

	block cipher.Block
}

func NewAESTool(key string) *AESTool {
	if n := len(key); 0 < n {
		var _key [aes.BlockSize * 16]byte
		//
		n = copy(_key[:], []byte(key))
		//
		if _n := n % aes.BlockSize; 0 != _n {
			n = ((n / aes.BlockSize) + 1) * aes.BlockSize
		}
		//
		if block, err := aes.NewCipher(_key[:n]); nil == err {
			//
			return &AESTool{
				block: block,
			}
		}
	}
	//
	return nil
}

func (this *AESTool) EncryptInit() {
	//
	this.Buffer.Reset()
	//
	io.CopyN(&this.Buffer, rand.Reader, aes.BlockSize)
}

func (this *AESTool) Encrypt(data []byte) error {
	if _, err := this.Buffer.Write(data); nil == err {
		//
		data := this.Buffer.Bytes()
		//
		stream := cipher.NewCFBEncrypter(this.block, data[:aes.BlockSize])
		//
		stream.XORKeyStream(data[aes.BlockSize:], data[aes.BlockSize:])
		//
		return nil
	} else {
		return err
	}
}

func (this *AESTool) Decrypt(data []byte) error {
	if _, err := this.Buffer.Write(data); nil == err {
		//
		data := this.Buffer.Bytes()
		//
		stream := cipher.NewCFBDecrypter(this.block, data[:aes.BlockSize])
		//
		stream.XORKeyStream(data[aes.BlockSize:], data[aes.BlockSize:])
		//
		this.Buffer.Next(aes.BlockSize)
		//
		return nil
	} else {
		return err
	}
}

func (this *AESTool) WriteByHexString(data string) error {
	if _data, err := hex.DecodeString(data); nil == err {
		if _, err = this.Buffer.Write(_data); nil == err {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

func (this *AESTool) ReadByHexString() string {
	return hex.EncodeToString(this.Buffer.Bytes())
}

func (this *AESTool) WriteByBase64(data string) error {
	if _data, err := base64.StdEncoding.DecodeString(data); nil == err {
		if _, err = this.Buffer.Write(_data); nil == err {
			return nil
		} else {
			return err
		}
	} else {
		return err
	}
}

func (this *AESTool) ReadByBase64() string {
	return base64.StdEncoding.EncodeToString(this.Buffer.Bytes())
}
