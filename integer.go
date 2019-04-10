package phpserialize

import (
	"errors"
	"reflect"
	"strconv"
)

var (
	integerPrefix = []byte("i:")
	integerSuffix = []byte(";")
)

func isInteger(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func (e *Encoder) encodeInteger(v reflect.Value) error {
	e.Write(integerPrefix)
	var b []byte
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		b = appendInt(e.scratch[:0], v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b = appendUint(e.scratch[:0], v.Uint())
	default:
		return TypeError{errors.New("%v is not an integer")}
	}
	b = append(b, ';')
	e.Write(b)
	return nil
}

func appendInt(dst []byte, i int64) []byte {
	return strconv.AppendInt(dst, i, 10)
}

func appendUint(dst []byte, i uint64) []byte {
	return strconv.AppendUint(dst, i, 10)
}
