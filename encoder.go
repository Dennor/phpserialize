// Package phpserialize is implementation of php serialize function
// based on github.com/mitsuhiko/phpserialize
package phpserialize

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sync"
)

type encoderPool struct {
	sync.Pool
}

func (e *encoderPool) Get() *Encoder {
	enc := e.Pool.Get().(*Encoder)
	enc.Reset()
	enc.w = nil
	return enc
}

func (e *encoderPool) Put(enc *Encoder) {
	e.Pool.Put(enc)
}

func newEncoderPool() *encoderPool {
	return &encoderPool{
		sync.Pool{
			New: func() interface{} {
				return &Encoder{}
			},
		},
	}
}

var encoderStatePool = newEncoderPool()

// Encoder implementing php serialize functionality
type Encoder struct {
	bytes.Buffer
	scratch scratchBuffer
	w       io.Writer
}

type encoderProp struct {
	key   func() error
	value reflect.Value
}

func (e *Encoder) encodeKey(v reflect.Value) error {
	switch {
	case isNil(v):
		e.WriteString(`s:0:"";`)
		return nil
	case isInteger(v):
		return e.encodeInteger(v)
	case isFloating(v):
		return e.encodeFloating(v)
	case isString(v):
		return e.encodeString(v)
	default:
		return TypeError{fmt.Errorf("can't serialize %v as key", v.Interface())}
	}
}

func (e *Encoder) encodeValue(v reflect.Value) error {
	if v.Type().Implements(marshalerType) {
		return e.encodeMarshaler(v)
	}
	if isNil(v) {
		return e.encodeNil(v)
	}
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	switch {
	case isBool(v):
		return e.encodeBool(v)
	case isInteger(v):
		return e.encodeInteger(v)
	case isFloating(v):
		return e.encodeFloating(v)
	case isString(v):
		return e.encodeString(v)
	case isIterable(v):
		return e.encodeIterable(v)
	case isMap(v):
		return e.encodeMap(v)
	case isStruct(v):
		return e.encodeStruct(v)
	case isPhpObject(v):
		return e.encodePhpObject(v)
	}
	return TypeError{fmt.Errorf("can't serialize %v", v.Interface())}
}

func (e *Encoder) encodePropsHeader(l int) error {
	e.WriteString("a:")
	e.Write(appendInt(e.scratch[:0], int64(l)))
	e.WriteString(":{")
	return nil
}

func (e *Encoder) encodePropsFinish() error {
	e.WriteByte('}')
	return nil
}

func (e *Encoder) encodeProp(prop encoderProp) error {
	if err := prop.key(); err != nil {
		return err
	}
	return e.encodeValue(prop.value)
}

// Encode value in php serialize format
func (e *Encoder) Encode(v interface{}) error {
	if err := e.encodeValue(reflect.ValueOf(v)); err != nil {
		return err
	}
	_, err := io.Copy(e.w, e)
	return err
}

// NewEncoder creates new encoder
func NewEncoder(w io.Writer) *Encoder {
	enc := encoderStatePool.Get()
	enc.w = w
	return enc
}

// Marshal v like Php serialize function
func Marshal(v interface{}) ([]byte, error) {
	enc := encoderStatePool.Get()
	if err := enc.encodeValue(reflect.ValueOf(v)); err != nil {
		return nil, err
	}
	b := append([]byte(nil), enc.Bytes()...)
	encoderStatePool.Put(enc)
	return b, nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
