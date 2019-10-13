package phpserialize

import (
	"reflect"
	"unicode/utf8"
)

const (
	quote = '\\'
)

var (
	zerob = [utf8.UTFMax * 4]byte{}
)

func isString(v reflect.Value) bool {
	return v.Kind() == reflect.String
}

func asciiPrintNonQuote(b byte) bool {
	return b >= ' ' && b <= '~' && b != '\\'
}

func (e *Encoder) encodeString(v reflect.Value) error {
	return e.encodeStringRaw(v.String())
}

func (e *Encoder) encodeStringRaw(s string) error {
	lengthChars := 1
	for len(s)/(10*lengthChars) > 0 {
		lengthChars++
	}
	e.Grow(4 + len(s) + lengthChars)
	e.Write([]byte{'s', ':'})
	last := len(appendInt(e.scratch[:0], int64(len(s))))
	e.Write(e.scratch[:last])
	e.Write([]byte{':', '"'})
	e.Write([]byte(s))
	e.Write([]byte{'"', ';'})
	return nil
}
