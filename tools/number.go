package tools

import (
	"errors"
	"fmt"
	"strconv"
)

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64 | ~complex64 | ~complex128
}

func NumberToString[T Number](num T) (string, error) {
	switch v := any(num).(type) {
	case int:
		return strconv.Itoa(v), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case complex64:
		return strconv.FormatComplex(complex128(v), 'f', -1, 64), nil
	case complex128:
		return strconv.FormatComplex(v, 'f', -1, 128), nil
	default:
		return "", errors.New("unsupported type: " + fmt.Sprintf("%T", v))
	}
}

func NumberToStringIgnoreError[T Number](num T) string {
	str, err := NumberToString(num)
	if err != nil {
		return ""
	}
	return str
}

func StringToNumber[T Number](str string) (T, error) {
	var zero T
	switch any(zero).(type) {
	case int:
		val, err := strconv.Atoi(str)
		if err != nil {
			return zero, err
		}
		return any(val).(T), nil
	case int8:
		val, err := strconv.ParseInt(str, 10, 8)
		if err != nil {
			return zero, err
		}
		return any(int8(val)).(T), nil
	case int16:
		val, err := strconv.ParseInt(str, 10, 16)
		if err != nil {
			return zero, err
		}
		return any(int16(val)).(T), nil
	case int32:
		val, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return zero, err
		}
		return any(int32(val)).(T), nil
	case int64:
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return zero, err
		}
		return any(val).(T), nil
	case uint:
		val, err := strconv.ParseUint(str, 10, 0)
		if err != nil {
			return zero, err
		}
		return any(uint(val)).(T), nil
	case uint8:
		val, err := strconv.ParseUint(str, 10, 8)
		if err != nil {
			return zero, err
		}
		return any(uint8(val)).(T), nil
	case uint16:
		val, err := strconv.ParseUint(str, 10, 16)
		if err != nil {
			return zero, err
		}
		return any(uint16(val)).(T), nil
	case uint32:
		val, err := strconv.ParseUint(str, 10, 32)
		if err != nil {
			return zero, err
		}
		return any(uint32(val)).(T), nil
	case uint64:
		val, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return zero, err
		}
		return any(val).(T), nil
	case float32:
		val, err := strconv.ParseFloat(str, 32)
		if err != nil {
			return zero, err
		}
		return any(float32(val)).(T), nil
	case float64:
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return zero, err
		}
		return any(val).(T), nil
	case complex64:
		val, err := strconv.ParseComplex(str, 64)
		if err != nil {
			return zero, err
		}
		return any(complex64(val)).(T), nil
	case complex128:
		val, err := strconv.ParseComplex(str, 128)
		if err != nil {
			return zero, err
		}
		return any(val).(T), nil
	default:
		return zero, errors.New("unsupported type: " + fmt.Sprintf("%T", zero))
	}
}

func StringToNumberIgnoreError[T Number](str string) T {
	num, err := StringToNumber[T](str)
	if err != nil {
		return 0
	}
	return num
}
