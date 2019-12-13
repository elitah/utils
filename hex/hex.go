package hex

import (
	"encoding/hex"
	"errors"
	"strings"
	"unicode"
)

const hextable = "0123456789abcdef"

var (
	empty = errors.New("Empty string")
)

func Encode(dst, src []byte, sep ...byte) int {
	j := 0
	max := len(dst) - 2
	for _, v := range src {
		dst[j] = hextable[v>>4]
		dst[j+1] = hextable[v&0x0f]
		if max > j && 0 < len(sep) && 0 != sep[0] {
			dst[j+2] = sep[0]
			j += 3
		} else {
			j += 2
		}
	}
	return len(src) * 2
}

func EncodeToString(src []byte, sep ...byte) string {
	if 0 < len(src) {
		var dst []byte
		if 2 <= len(src) && 0 < len(sep) && 0 != sep[0] {
			dst = make([]byte, hex.EncodedLen(len(src))+len(src)-1)
			Encode(dst, src, sep[0])
		} else {
			dst = make([]byte, hex.EncodedLen(len(src)))
			Encode(dst, src)
		}
		return string(dst)
	}
	return ""
}

func EncodeToStringWithSeq(src []byte, sep rune) string {
	if s := byte(sep); 0x20 <= s && 0x7A >= s {
		return EncodeToString(src, s)
	}
	return EncodeToString(src)
}

func DecodeStringWithSeq(s string) ([]byte, error) {
	list := strings.FieldsFunc(s, func(r rune) bool {
		if unicode.IsNumber(r) {
			return false
		} else if 'a' <= r && 'f' >= r {
			return false
		} else if 'A' <= r && 'F' >= r {
			return false
		}
		return true
	})
	if 0 < len(list) {
		s = strings.Join(list, "")
		return hex.DecodeString(s)
	}
	return nil, empty
}
