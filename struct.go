package phpserialize

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type Stringer interface {
	String() string
}

var (
	goStringerType = reflect.TypeOf((*fmt.GoStringer)(nil)).Elem()
	stringerType   = reflect.TypeOf((*Stringer)(nil)).Elem()
)

func isStruct(v reflect.Value) bool {
	return v.Kind() == reflect.Struct
}

func getTag(field *reflect.StructField) (string, []string) {
	tag := field.Tag.Get("php")
	if tag == "" {
		tag = field.Tag.Get("json")
		if tag == "" {
			return "", nil
		}
	}
	parts := strings.Split(tag, ",")
	return parts[0], parts[1:]
}

func omitempty(opts []string) bool {
	for _, opt := range opts {
		if opt == "omitempty" {
			return true
		}
	}
	return false
}

func asString(opts []string) bool {
	for _, opt := range opts {
		if opt == "string" {
			return true
		}
	}
	return false
}

type field struct {
	typ        reflect.Type
	tagged     bool
	name       string
	omitEmpty  bool
	asString   bool
	index      []int
	encode     func(e *Encoder, v reflect.Value) error
	encodedKey string
}

func (e *Encoder) encodeStruct(v reflect.Value) error {
	fields := cachedTypeFields(v.Type())
	var fieldsCount int
	senc := encoderStatePool.Get()
	for i := 0; i < len(fields); i++ {
		fv := v.Field(fields[i].index[0])
		for _, idx := range fields[i].index[1:] {
			if fv.Kind() == reflect.Ptr {
				if fv.IsNil() {
					fv.Set(reflect.New(fv.Type().Elem()))
				}
				fv = fv.Elem()
			}
			fv = fv.Field(idx)
		}
		if fields[i].omitEmpty && isEmptyValue(fv) {
			continue
		}
		fieldsCount++
		senc.WriteString(fields[i].encodedKey)
		fields[i].encode(senc, fv)
	}
	e.encodePropsHeader(fieldsCount)
	e.Write(senc.Bytes())
	e.encodePropsFinish()
	encoderStatePool.Put(senc)
	return nil
}

var fieldCache sync.Map

func typeFields(t reflect.Type) []field {
	current := []field{}
	next := []field{{typ: t}}
	visited := map[reflect.Type]bool{}
	fieldAt := map[string]int{}
	orphans := []int{}
	var fields []field
	var level int
	for len(next) > 0 {
		level++
		current, next = next, current[:0]
		nextCount := map[reflect.Type]bool{}
		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				isUnexported := sf.PkgPath != ""
				if sf.Anonymous {
					t := sf.Type
					if t.Kind() == reflect.Ptr {
						t = t.Elem()
					}
					if isUnexported && t.Kind() != reflect.Struct {
						continue
					}
				} else if isUnexported {
					continue
				}
				tag, opts := getTag(&sf)
				if tag == "-" {
					continue
				}
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i
				name := tag
				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					if fAt, ok := fieldAt[name]; ok {
						if level > len(fields[fAt].index) {
							continue
						}
						if fields[fAt].tagged || (!tagged && !fields[fAt].tagged) {
							continue
						}
						orphans = append(orphans, fAt)
					}
					fieldAt[name] = len(fields)
					fields = append(fields, field{
						typ:       ft,
						tagged:    tagged,
						name:      name,
						omitEmpty: omitempty(opts),
						asString:  asString(opts),
						index:     index,
					})
					continue
				}
				if !nextCount[ft] {
					nextCount[ft] = true
					next = append(next, field{index: index, typ: ft})
				}
			}
		}
	}
	for i, orphan := range orphans {
		fields = append(fields[:orphan-i], fields[orphan-i+1:]...)
	}
	for i := range fields {
		typ := t
		for _, i := range fields[i].index {
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			typ = typ.Field(i).Type
		}
		if fields[i].asString {
			fields[i].encode = asStringEncoder(typ)
		} else {
			fields[i].encode = typeEncoder(typ)
		}
		fields[i].encodedKey = encodeStructKey(fields[i].name)
	}
	return fields
}

func encodeStructKey(name string) string {
	keyBuf, _ := Marshal(name)
	return string(keyBuf)
}

func typeEncoder(t reflect.Type) func(*Encoder, reflect.Value) error {
	if t.Implements(marshalerType) {
		return func(e *Encoder, v reflect.Value) error {
			return e.encodeMarshaler(v)
		}
	}
	if t == reflect.TypeOf(PhpObject{}) {
		return func(e *Encoder, v reflect.Value) error {
			return e.encodePhpObject(v)
		}
	}
	switch t.Kind() {
	case reflect.Bool:
		return func(e *Encoder, v reflect.Value) error {
			return e.encodeBool(v)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func(e *Encoder, v reflect.Value) error {
			return e.encodeInteger(v)
		}
	case reflect.Float32, reflect.Float64:
		return func(e *Encoder, v reflect.Value) error {
			return e.encodeFloating(v)
		}
	case reflect.String:
		return func(e *Encoder, v reflect.Value) error {
			return e.encodeString(v)
		}
	case reflect.Slice, reflect.Array:
		return func(e *Encoder, v reflect.Value) error {
			return e.encodeIterable(v)
		}
	case reflect.Map:
		return func(e *Encoder, v reflect.Value) error {
			return e.encodeMap(v)
		}
	case reflect.Struct:
		return func(e *Encoder, v reflect.Value) error {
			return e.encodeStruct(v)
		}
	case reflect.Ptr:
		f := typeEncoder(t.Elem())
		return func(e *Encoder, v reflect.Value) error {
			if v.IsNil() {
				return e.encodeNil(v)
			}
			return f(e, v)
		}
	case reflect.Interface:
		return func(e *Encoder, v reflect.Value) error {
			if v.IsNil() {
				return e.encodeNil(v)
			}
			return e.encodeValue(v)
		}
	}
	return nil
}

func asStringEncoder(t reflect.Type) func(e *Encoder, v reflect.Value) error {
	if t.Implements(goStringerType) {
		return func(e *Encoder, v reflect.Value) error {
			return e.encodeStringRaw(v.Interface().(fmt.GoStringer).GoString())
		}
	}
	if t.Implements(stringerType) {
		return func(e *Encoder, v reflect.Value) error {
			return e.encodeStringRaw(v.Interface().(Stringer).String())
		}
	}
	switch t.Kind() {
	case reflect.Bool:
		return func(e *Encoder, v reflect.Value) error {
			if v.Bool() {
				return e.encodeStringRaw(`s:1:"true";`)
			}
			return e.encodeStringRaw(`s:1:"false";`)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func(e *Encoder, v reflect.Value) error {
			b := appendInt(e.scratch[:0], v.Int())
			lb := appendInt(e.scratch[len(b):len(b)], int64(len(b)))
			e.WriteString(`s:`)
			e.Write(e.scratch[len(b) : len(b)+len(lb)])
			e.WriteString(`:"`)
			e.Write(e.scratch[:len(b)])
			e.WriteString(`";`)
			return nil
		}
	case reflect.Float32, reflect.Float64:
		return func(e *Encoder, v reflect.Value) error {
			bitSize := 32
			if t.Kind() == reflect.Float64 {
				bitSize = 64
			}
			b := appendFloat(e.scratch[:0], v.Float(), bitSize)
			lb := appendInt(e.scratch[len(b):len(b)], int64(len(b)))
			e.WriteString(`s:`)
			e.Write(e.scratch[len(b):len(lb)])
			e.WriteString(`:"`)
			e.Write(e.scratch[:len(b)])
			e.WriteString(`";`)
			return nil
		}
	case reflect.String:
		return func(e *Encoder, v reflect.Value) error {
			return e.encodeStringRaw(v.String())
		}
	case reflect.Ptr:
		f := asStringEncoder(t.Elem())
		if f == nil {
			return nil
		}
		return func(e *Encoder, v reflect.Value) error {
			if v.IsNil() {
				return e.encodeNil(v)
			}
			return f(e, v)
		}
	}
	return nil
}

func cachedTypeFields(t reflect.Type) []field {
	if f, ok := fieldCache.Load(t); ok {
		return f.([]field)
	}
	f, _ := fieldCache.LoadOrStore(t, typeFields(t))
	return f.([]field)
}
