package phpserialize

import (
	"reflect"
)

func isIterable(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}

func (e *Encoder) encodeIterable(v reflect.Value) error {
	l := v.Len()
	e.encodePropsHeader(l)
	for i := 0; i < v.Len(); i++ {
		copy(e.scratch[:], integerPrefix)
		last := len(appendInt(e.scratch[:len(integerPrefix)], int64(i)))
		e.scratch[last] = ';'
		e.Write(e.scratch[:last+1])
		e.encodeValue(v.Index(i))
	}
	e.encodePropsFinish()
	return nil
}
