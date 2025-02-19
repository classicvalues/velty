package est

import (
	"github.com/viant/xunsafe"
	"reflect"
)

const defaultPkg = "github.com/viant/velty/est"

//Type represents scope type
type Type struct {
	reflect.Type
	types   map[string]int
	fields  []reflect.StructField
	xFields []*xunsafe.Field
}

func (t *Type) AddField(id string, name string, rType reflect.Type) reflect.StructField {
	return t.addField(id, name, rType, false)
}

func (t *Type) EmbedType(id string, name string, rType reflect.Type) reflect.StructField {
	field := t.addField(id, name, rType, true)
	return field
}

func (t *Type) addField(id string, name string, rType reflect.Type, anonymous bool) reflect.StructField {
	pkg := ""
	if name[0] > 'Z' {
		pkg = defaultPkg
	}

	field := reflect.StructField{Name: name, Type: rType, PkgPath: pkg, Anonymous: anonymous}

	offset := uintptr(0)
	for _, structField := range t.fields {
		size := structField.Type.Size()
		offset += size
	}

	t.fields = append(t.fields, field)
	t.Type = reflect.StructOf(t.fields)
	field.Offset = offset

	t.xFields = append(t.xFields, xunsafe.NewField(field))

	t.types[id] = len(t.fields) - 1
	return field
}

func (t *Type) Mutator(id string) (*xunsafe.Field, bool) {
	index, found := t.types[id]
	if !found {
		return nil, false
	}

	return t.xFields[index], true
}

func NewType() *Type {
	return &Type{
		Type:  reflect.StructOf([]reflect.StructField{}),
		types: map[string]int{},
	}
}
