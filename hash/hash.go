package hash

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/gob"
	"fmt"
	"hash"
	"io"
)

var (
	gobFormat bool = false
)

func SetGobFormat(flag bool) {
	gobFormat = flag
}

func New(name string) hash.Hash {
	switch name {
	case "md5":
		return md5.New()
	case "sha1":
		return sha1.New()
	case "sha256":
		return sha256.New()
	case "sha512":
		return sha512.New()
	}
	return nil
}

func Write(h hash.Hash, args ...interface{}) (int, error) {
	var b bytes.Buffer

	enc := gob.NewEncoder(&b)

	for _, arg := range args {
		switch result := arg.(type) {
		case string:
			if _, err := b.WriteString(result); nil != err {
				return 0, err
			}
		case []byte:
			if _, err := b.Write(result); nil != err {
				return 0, err
			}
		case byte:
			if err := b.WriteByte(result); nil != err {
				return 0, err
			}
		default:
			if data, ok := arg.(interface {
				Bytes() []byte
			}); ok {
				if _, err := b.Write(data.Bytes()); nil != err {
					return 0, err
				}
			} else {
				if gobFormat {
					if err := enc.Encode(arg); nil != err {
						return 0, err
					}
				} else {
					if _, err := b.WriteString(fmt.Sprint(arg)); nil != err {
						return 0, err
					}
				}
			}
		}
	}

	if 0 < b.Len() {
		return h.Write(b.Bytes())
	}

	return 0, fmt.Errorf("empty write")
}

func WriteString(h hash.Hash, args ...string) {
	if nil != h {
		for _, item := range args {
			io.WriteString(h, item)
		}
	}
}

func SumString(h hash.Hash) string {
	return fmt.Sprintf("%x", h.Sum(nil))
}

func HashToBytes(name string, args ...interface{}) []byte {
	if h := New(name); nil != h {
		Write(h, args...)
		return h.Sum(nil)[:]
	}
	return nil
}

func HashToString(name string, args ...interface{}) string {
	if h := New(name); nil != h {
		Write(h, args...)
		return SumString(h)
	}
	return ""
}
