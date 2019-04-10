package phpserialize

import (
	"reflect"
)

type PhpObject struct {
	Name    string
	PhpVars interface{}
}

func isPhpObject(v reflect.Value) bool {
	_, ok := v.Interface().(PhpObject)
	return ok
}

func (e *Encoder) encodePhpObject(v reflect.Value) error {
	phpObject := v.Interface().(PhpObject)
	e.WriteString("O:")
	e.Write(appendInt(e.scratch[:0], int64(len(phpObject.Name))))
	prefixAt := e.Len()
	if err := e.encodeValue(reflect.ValueOf(phpObject.PhpVars)); err != nil {
		return err
	}
	// replace prefix form vars with :
	e.Bytes()[prefixAt] = ':'
	e.WriteByte('}')
	return nil
}
