package random

import (
	"crypto/rand"
	"errors"
	"strings"
	"sync"
	"time"
)

const (
	ModeALL           = iota // ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_
	ModeNoLower              // ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_
	ModeNoUpper              // abcdefghijklmnopqrstuvwxyz0123456789-_
	ModeNoNumber             // ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_
	ModeNoLowerNumber        // ABCDEFGHIJKLMNOPQRSTUVWXYZ-_
	ModeNoUpperNumber        // abcdefghijklmnopqrstuvwxyz-_
	ModeNoLine               // ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789
	ModeNoLowerLine          // ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789
	ModeNoUpperLine          // abcdefghijklmnopqrstuvwxyz0123456789
	ModeOnlyUpper            // ABCDEFGHIJKLMNOPQRSTUVWXYZ
	ModeOnlyLower            // abcdefghijklmnopqrstuvwxyz
	ModeOnlyNumber           // 0123456789
	ModeHexUpper             // 0123456789ABCDEF
	ModeHexLower             // 0123456789abcdef

	ModeLimit
)

var (
	ENoResult = errors.New("Unable read result")

	pool = sync.Pool{
		New: func() interface{} {
			return &fastRandomString{}
		},
	}

	charsUpper    = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ") // 大写字母
	charsLower    = []byte("abcdefghijklmnopqrstuvwxyz") // 小写字母
	charsLine     = []byte("-_")                         // 中/下划线
	charsNumber   = []byte("0123456789")                 // 数字
	charsHexUpper = []byte("0123456789ABCDEF")           // 十六进制(大写字母)
	charsHexLower = []byte("0123456789abcdef")           // 十六进制(小写字母)
)

type fastRandomString struct {
	strings.Builder

	tables [64]byte
	rdbuf  [2056]byte
}

func (this *fastRandomString) Update(mode, max int) bool {
	if ModeALL <= mode && ModeLimit > mode && 2 <= max {
		if 2048 < max {
			max = 2048
		}
		if n, err := rand.Read(this.rdbuf[:max]); nil == err {
			if max <= n {
				var length int
				switch mode {
				case ModeALL: // ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_
					length += copy(this.tables[0:26], charsUpper)
					length += copy(this.tables[26:52], charsLower)
					length += copy(this.tables[52:62], charsNumber)
					length += copy(this.tables[62:64], charsLine)
				case ModeNoLower: // ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_
					length += copy(this.tables[0:26], charsUpper)
					length += copy(this.tables[26:36], charsNumber)
					length += copy(this.tables[36:38], charsLine)
				case ModeNoUpper: // abcdefghijklmnopqrstuvwxyz0123456789-_
					length += copy(this.tables[0:26], charsLower)
					length += copy(this.tables[26:36], charsNumber)
					length += copy(this.tables[36:38], charsLine)
				case ModeNoNumber: // ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_
					length += copy(this.tables[0:26], charsUpper)
					length += copy(this.tables[26:52], charsLower)
					length += copy(this.tables[52:54], charsLine)
				case ModeNoLowerNumber: // ABCDEFGHIJKLMNOPQRSTUVWXYZ-_
					length += copy(this.tables[0:26], charsUpper)
					length += copy(this.tables[26:28], charsLine)
				case ModeNoUpperNumber: // abcdefghijklmnopqrstuvwxyz-_
					length += copy(this.tables[0:26], charsLower)
					length += copy(this.tables[26:28], charsLine)
				case ModeNoLine: // ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789
					length += copy(this.tables[0:26], charsUpper)
					length += copy(this.tables[26:52], charsLower)
					length += copy(this.tables[52:62], charsNumber)
				case ModeNoLowerLine: // ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789
					length += copy(this.tables[0:26], charsUpper)
					length += copy(this.tables[26:36], charsNumber)
				case ModeNoUpperLine: // abcdefghijklmnopqrstuvwxyz0123456789
					length += copy(this.tables[0:26], charsLower)
					length += copy(this.tables[26:36], charsNumber)
				case ModeOnlyUpper: // ABCDEFGHIJKLMNOPQRSTUVWXYZ
					length += copy(this.tables[0:26], charsUpper)
				case ModeOnlyLower: // abcdefghijklmnopqrstuvwxyz
					length += copy(this.tables[0:26], charsLower)
				case ModeOnlyNumber: // 0123456789
					length += copy(this.tables[0:10], charsNumber)
				case ModeHexUpper: // 0123456789ABCDEF
					length += copy(this.tables[0:16], charsHexUpper)
				case ModeHexLower: // 0123456789abcdef
					length += copy(this.tables[0:16], charsHexLower)
				default:
					return false
				}

				if 0 < length {
					var offset int

					//fmt.Printf("table: \033[31;1m%s\033[0m\n", string(this.tables[:length]))

					this.Reset()

					for i := 0; max > i; i++ {
						offset = int(this.rdbuf[i]) % length
						if 0 == i {
							for '-' == this.tables[offset] || '_' == this.tables[offset] {
								offset = (int(time.Now().UnixNano()) + offset) % length
							}
						}
						this.WriteByte(this.tables[offset])
					}

					return 0 < this.Len()
				}
			}
		}
	}

	return false
}

func NewRandomString(mode, max int) (result string) {
	if 2 <= max {
		if fr, ok := pool.Get().(*fastRandomString); ok {
			if fr.Update(mode, max) {
				result = fr.String()
				pool.Put(fr)
				return
			}
		}
	}
	panic(ENoResult)
}
