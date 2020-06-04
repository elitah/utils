package hex

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"strconv"
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

func EncodeNumberToStringWithSeq(v interface{}, sep rune, le bool, n ...int) string {
	var value uint64

	switch result := v.(type) {
	case string:
		if s, err := strconv.ParseUint(result, 10, 64); err == nil {
			value = s
		}
	case int8, uint8:
		return fmt.Sprintf("%02X", result)
	case int16:
		return fmt.Sprintf("%02X %02X", (uint16(math.Abs(float64(result)))>>8)&0xFF, uint16(math.Abs(float64(result)))&0xFF)
	case uint16:
		return fmt.Sprintf("%02X %02X", (result>>8)&0xFF, result&0xFF)
	case int:
		value = uint64(math.Abs(float64(result)))
	case int32:
		value = uint64(math.Abs(float64(result)))
	case uint32:
		value = uint64(result)
	case int64:
		value = uint64(math.Abs(float64(result)))
	case uint64:
		value = uint64(result)
	}

	// 修正max
	max := bits.Len64(value)/8 + 1

	if 0 < len(n) && 0 < n[0] && max != n[0] {
		max = n[0]
	}

	data := make([]byte, max)

	for i, _ := range data {
		if le {
			data[i] = byte(value >> (i * 8))
		} else {
			data[i] = byte(value >> ((max - i - 1) * 8))
		}
	}

	return EncodeToStringWithSeq(data, sep)
}

func DecodeStringWithSeq(s string) ([]byte, error) {
	list := strings.FieldsFunc(s, func(r rune) bool {
		if unicode.IsNumber(r) {
			return false
		} else if unicode.IsLetter(r) {
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
