package phpserialize

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhpSerialize(t *testing.T) {
	assert := assert.New(t)
	type PhpTagged struct {
		AField int `php:"a_field"`
		BField int `php:"b_field,omitempty"`
	}
	type JsonTagged struct {
		AField int `json:"a_field"`
		BField int `json:"b_field,omitempty"`
	}
	data := []struct {
		v        interface{}
		expected string
	}{
		{5, "i:5;"},
		{5.6, "d:5.6;"},
		{"Hello world", `s:11:"Hello world";`},
		{"Björk Guðmundsdóttir", `s:23:"Bj\xc3\xb6rk Gu\xc3\xb0mundsd\xc3\xb3ttir";`},
		{`Hello
world`, `s:11:"Hello\nworld";`},
		{"\001\002\003", `s:3:"\x01\x02\x03";`},
		{[]int{7, 8, 9}, "a:3:{i:0;i:7;i:1;i:8;i:2;i:9;}"},
		{PhpTagged{AField: 1}, `a:1:{s:7:"a_field";i:1;}`},
		{PhpTagged{AField: 1, BField: 2}, `a:2:{s:7:"a_field";i:1;s:7:"b_field";i:2;}`},
		{JsonTagged{AField: 1}, `a:1:{s:7:"a_field";i:1;}`},
		{JsonTagged{AField: 1, BField: 2}, `a:2:{s:7:"a_field";i:1;s:7:"b_field";i:2;}`},
	}
	for _, tt := range data {
		b, err := Marshal(tt.v)
		assert.NoError(err)
		assert.Equal([]byte(tt.expected), b, "value %v, result: %s", tt.v, tt.expected)
	}

	m := map[string]interface{}{"a": 1, "b": 2, "c": 3}
	b, err := Marshal(m)
	ms := string(b)
	assert.NoError(err)
	assert.True(strings.HasPrefix(ms, "a:3:{"))
	assert.True(strings.HasSuffix(ms, "}"))
	assert.NotEqual(strings.Index(ms, `s:1:"a";i:1;`), -1)
	assert.NotEqual(strings.Index(ms, `s:1:"b";i:2;`), -1)
	assert.NotEqual(strings.Index(ms, `s:1:"c";i:3;`), -1)
	assert.Len(ms, 42)
}

var bb []byte

func benchmarkPhpSerialize(v interface{}, b *testing.B) {
	var br []byte
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		br, _ = Marshal(v)
	}
	bb = br
}

func benchmarkJsonSerialize(v interface{}, b *testing.B) {
	var br []byte
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		br, _ = json.Marshal(v)
	}
	bb = br
}

func BenchmarkPhpSerializeInteger(b *testing.B) {
	benchmarkPhpSerialize(5, b)
}

func BenchmarkJSONSerializeInteger(b *testing.B) {
	benchmarkJsonSerialize(5, b)
}

func BenchmarkPhpSerializeFloating(b *testing.B) {
	benchmarkPhpSerialize(5.6, b)
}

func BenchmarkJSONSerializeFloating(b *testing.B) {
	benchmarkJsonSerialize(5.6, b)
}

func BenchmarkPhpSerializeString(b *testing.B) {
	benchmarkPhpSerialize("Björk Guðmundsdóttir", b)
}

func BenchmarkJSONSerializeString(b *testing.B) {
	benchmarkJsonSerialize("Björk Guðmundsdóttir", b)
}

func BenchmarkPhpSerializeSlice(b *testing.B) {
	benchmarkPhpSerialize([]int{7, 8, 9}, b)
}

func BenchmarkJSONSerializeSlice(b *testing.B) {
	benchmarkJsonSerialize([]int{7, 8, 9}, b)
}

func BenchmarkPhpSerializeStruct(b *testing.B) {
	benchmarkPhpSerialize(struct {
		AField int `php:"a_field"`
		BField int `php:"b_field"`
		CField int `php:"c_field"`
		DField int `php:"d_field"`
		EField int `php:"e_field"`
	}{AField: 1, BField: 2}, b)
}

func BenchmarkJSONSerializeStruct(b *testing.B) {
	benchmarkJsonSerialize(struct {
		AField int `json:"a_field"`
		BField int `json:"b_field"`
		CField int `json:"c_field"`
		DField int `json:"d_field"`
		EField int `json:"e_field"`
	}{AField: 1, BField: 2}, b)
}

func BenchmarkPhpSerializeMap(b *testing.B) {
	benchmarkPhpSerialize(map[string]interface{}{"a": 1, "b": 2, "c": 3}, b)
}

func BenchmarkJSONSerializeMap(b *testing.B) {
	benchmarkJsonSerialize(map[string]interface{}{"a": 1, "b": 2, "c": 3}, b)
}
