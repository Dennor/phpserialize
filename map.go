package phpserialize

import (
	"reflect"
)

func isMap(v reflect.Value) bool {
	return v.Kind() == reflect.Map
}

func (e *Encoder) encodeMap(v reflect.Value) error {
	mks := v.MapKeys()
	l := len(mks)
	e.encodePropsHeader(l)
	for _, mk := range mks {
		e.encodeProp(encoderProp{
			key: func() error {
				return e.encodeKey(mk)
			},
			value: v.MapIndex(mk),
		})
	}
	e.encodePropsFinish()
	return nil
}
