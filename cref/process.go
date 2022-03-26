package cref

import (
	"errors"
	"reflect"
)

type CachedType struct {
	Fields map[string]reflect.Type
}

func ComputeCachedType(t reflect.Type) (res *CachedType, err error) {
	if t.Kind() != reflect.Struct {
		err = errors.New("arcah/cref: type must be struct")
		return
	}

	ires := CachedType{
		Fields: map[string]reflect.Type{},
	}

	length := t.NumField()
	for i := 0; i < length; i++ {
		f := t.Field(i)
		ires.Fields[f.Name] = f.Type
	}

	res = &ires
	return
}
