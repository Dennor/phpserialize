package phpserialize

import "reflect"

type Marshaler interface {
	MarshalPHP() ([]byte, error)
}

var (
	marshalerType = reflect.TypeOf((*Marshaler)(nil)).Elem()
)

func (e *Encoder) encodeMarshaler(v reflect.Value) error {
	marshaler := v.Interface().(Marshaler)
	b, err := marshaler.MarshalPHP()
	if err != nil {
		return err
	}
	e.Write(b)
	return nil
}
