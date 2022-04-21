package refutil

import "reflect"

func ValueIsEmpty(target reflect.Value) bool {
	return target.IsZero()
}

func InterfaceIsEmpty(empty interface{}) bool {
	return ValueIsEmpty(reflect.ValueOf(empty))
}
