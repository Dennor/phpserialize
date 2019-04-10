package phpserialize

import (
	"reflect"
	"strconv"
)

func isFloating(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func (e *Encoder) encodeFloating(v reflect.Value) error {
	e.WriteString("d:")
	bitSize := 64
	if v.Kind() == reflect.Float32 {
		bitSize = 32
	}
	b := appendFloat(e.scratch[:0], v.Float(), bitSize)
	e.Write(b)
	e.WriteByte(';')
	return nil
}

func appendFloat(dst []byte, f float64, bitSize int) []byte {
	return strconv.AppendFloat(dst, f, 'f', -1, bitSize)
}
