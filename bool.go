package phpserialize

import (
	"reflect"
)

func isBool(v reflect.Value) bool {
	return v.Kind() == reflect.Bool
}

func (e *Encoder) encodeBool(v reflect.Value) error {
	b := e.scratch[:0]
	b = append(b, "b:0;"...)
	if v.Bool() {
		b[2] = '1'
	}
	e.Write(b)
	return nil
}
