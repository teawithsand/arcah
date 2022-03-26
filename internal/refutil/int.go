package refutil

import "reflect"

func ValueToNumber(v reflect.Value) (res interface{}) {
	if v.Kind() == reflect.Int || v.Kind() == reflect.Int8 || v.Kind() == reflect.Int16 || v.Kind() == reflect.Int32 || v.Kind() == reflect.Int64 {
		return v.Int()
	} else if v.Kind() == reflect.Uint || v.Kind() == reflect.Uint16 || v.Kind() == reflect.Uint32 || v.Kind() == reflect.Uint64 {
		return v.Uint()
	} else if v.Kind() == reflect.Float32 || v.Kind() == reflect.Float64 {
		return v.Float()
	}
	return
}

func ValueIsEmpty(target reflect.Value) bool {
	return target.IsZero()
}

func InterfaceIsEmpty(empty interface{}) bool {
	return ValueIsEmpty(reflect.ValueOf(empty))
}
