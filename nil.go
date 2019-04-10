package phpserialize

import "reflect"

func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

var (
	phpNil = []byte("N;")
)

func (e *Encoder) encodeNil(v reflect.Value) error {
	e.WriteString("N;")
	return nil
}
