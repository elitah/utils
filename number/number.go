package number

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ENOSUPPORT = errors.New("no support")
)

func IsNumeric(x interface{}) bool {
	if nil != x {
		switch x.(type) {
		// using default
		//case bool:
		// using default
		//case string:
		case int, int8, int16, int32, int64:
			return true
		case uint, uint8, uint16, uint32, uint64, uintptr:
			return true
		// alias for uint8
		//case byte:
		//	return true
		// alias for int32
		//case rune:
		//	return true
		case float32, float64:
			return true
		case complex64, complex128:
			return true
		}
	}
	return false
}

func ToInt64(x interface{}) (int64, error) {
	if nil != x {
		switch result := x.(type) {
		case bool:
			if result {
				return 1, nil
			}
			return 0, nil
		case string:
			return strconv.ParseInt(result, 0, 64)
		case int:
			return int64(result), nil
		case int8:
			return int64(result), nil
		case int16:
			return int64(result), nil
		case int32:
			return int64(result), nil
		case int64:
			return int64(result), nil
		case uint:
			return int64(result), nil
		case uint8:
			return int64(result), nil
		case uint16:
			return int64(result), nil
		case uint32:
			return int64(result), nil
		case uint64:
			return int64(result), nil
		case uintptr:
			return int64(result), nil
		// alias for uint8
		//case byte:
		//	return int64(result), nil
		// alias for int32
		//case rune:
		//	return int64(result), nil
		case float32:
			return int64(result), nil
		case float64:
			return int64(result), nil
		// not support complex
		//case complex64:
		//	return int64(result), nil
		//case complex128:
		//	return int64(result), nil
		case interface{ Int64() (int64, error) }:
			return result.Int64()
		case interface{ Uint64() (uint64, error) }:
			if v, err := result.Uint64(); nil == err {
				return int64(v), nil
			} else {
				return 0, err
			}
		case interface{ Float64() (float64, error) }:
			if v, err := result.Float64(); nil == err {
				return int64(v), nil
			} else {
				return 0, err
			}
		default:
			return strconv.ParseInt(fmt.Sprint(result), 0, 64)
		}
	}
	return 0, ENOSUPPORT
}

func ToUint64(x interface{}) (uint64, error) {
	if nil != x {
		switch result := x.(type) {
		case bool:
			if result {
				return 1, nil
			}
			return 0, nil
		case string:
			return strconv.ParseUint(result, 0, 64)
		case int:
			return uint64(result), nil
		case int8:
			return uint64(result), nil
		case int16:
			return uint64(result), nil
		case int32:
			return uint64(result), nil
		case int64:
			return uint64(result), nil
		case uint:
			return uint64(result), nil
		case uint8:
			return uint64(result), nil
		case uint16:
			return uint64(result), nil
		case uint32:
			return uint64(result), nil
		case uint64:
			return uint64(result), nil
		case uintptr:
			return uint64(result), nil
		// alias for uint8
		//case byte:
		//	return uint64(result), nil
		// alias for int32
		//case rune:
		//	return uint64(result), nil
		case float32:
			return uint64(result), nil
		case float64:
			return uint64(result), nil
		// not support complex
		//case complex64:
		//	return uint64(result), nil
		//case complex128:
		//	return uint64(result), nil
		case interface{ Int64() (int64, error) }:
			if v, err := result.Int64(); nil == err {
				return uint64(v), nil
			} else {
				return 0, err
			}
		case interface{ Uint64() (uint64, error) }:
			return result.Uint64()
		case interface{ Float64() (float64, error) }:
			if v, err := result.Float64(); nil == err {
				return uint64(v), nil
			} else {
				return 0, err
			}
		default:
			return strconv.ParseUint(fmt.Sprint(result), 0, 64)
		}
	}
	return 0, ENOSUPPORT
}

func ToFloat64(x interface{}) (float64, error) {
	if nil != x {
		switch result := x.(type) {
		case bool:
			if result {
				return 1, nil
			}
			return 0, nil
		case string:
			return strconv.ParseFloat(result, 64)
		case int:
			return float64(result), nil
		case int8:
			return float64(result), nil
		case int16:
			return float64(result), nil
		case int32:
			return float64(result), nil
		case int64:
			return float64(result), nil
		case uint:
			return float64(result), nil
		case uint8:
			return float64(result), nil
		case uint16:
			return float64(result), nil
		case uint32:
			return float64(result), nil
		case uint64:
			return float64(result), nil
		case uintptr:
			return float64(result), nil
		// alias for uint8
		//case byte:
		//	return float64(result), nil
		// alias for int32
		//case rune:
		//	return float64(result), nil
		case float32:
			return float64(result), nil
		case float64:
			return float64(result), nil
		// not support complex
		//case complex64:
		//	return float64(result), nil
		//case complex128:
		//	return float64(result), nil
		case interface{ Int64() (int64, error) }:
			if v, err := result.Int64(); nil == err {
				return float64(v), nil
			} else {
				return 0, err
			}
		case interface{ Uint64() (uint64, error) }:
			if v, err := result.Uint64(); nil == err {
				return float64(v), nil
			} else {
				return 0, err
			}
		case interface{ Float64() (float64, error) }:
			return result.Float64()
		default:
			return strconv.ParseFloat(fmt.Sprint(result), 64)
		}
	}
	return 0, ENOSUPPORT
}
