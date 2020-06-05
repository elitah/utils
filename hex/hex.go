package hex

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode"
)

const hextable_l = "0123456789abcdef"
const hextable_u = "0123456789ABCDEF"

var (
	lowercase = uint32(1)

	empty = errors.New("Empty string")
)

func OutputLowerCase(enabled bool) {
	if enabled {
		atomic.StoreUint32(&lowercase, 1)
		return
	}
	atomic.StoreUint32(&lowercase, 0)
}

func EncodedLenWithSeq(n int) int {
	return (n * 3) - 1
}

func Encode(dst, src []byte, sep ...byte) (n int) {
	if 0 == len(sep) || 2 > len(src) {
		if hex.EncodedLen(len(src)) <= len(dst) {
			return hex.Encode(dst, src)
		}
		return 0
	}
	if 0x20 <= sep[0] && 0x7A >= sep[0] {
		if _n := (len(dst) + 1) / 3; 0 < _n {
			//
			table := []byte(hextable_u)
			//
			if 0x0 == atomic.LoadUint32(&lowercase) {
				table = []byte(hextable_l)
			}
			//
			if _n < len(src) {
				src = src[:_n]
			}
			//
			for i, item := range src {
				if 0 < i {
					dst[n] = sep[0]
					n++
				}
				dst[n] = table[item>>4]
				dst[n+1] = table[item&0x0f]
				n += 2
			}
		}
	}
	return
}

func EncodeToString(src []byte, sep ...byte) string {
	if 0 < len(src) {
		var s strings.Builder
		var r byte
		//
		table := []byte(hextable_u)
		//
		if 0x0 == atomic.LoadUint32(&lowercase) {
			table = []byte(hextable_l)
		}
		//
		if 0 < len(sep) && 0x20 <= sep[0] && 0x7A >= sep[0] {
			r = sep[0]
		}
		//
		s.Grow(EncodedLenWithSeq(len(src)))
		//
		for i, item := range src {
			if 0 < r && 0 < i {
				s.WriteByte(r)
			}
			s.WriteByte(table[item>>4])
			s.WriteByte(table[item&0x0f])
		}
		//
		return s.String()
	}
	//
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
