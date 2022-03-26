package arcah

import (
	"errors"
	"reflect"

	"github.com/teawithsand/reval/stdesc"
)

type QueryReflector struct {
	DescriptorComputer  *stdesc.Comptuer
	ReflectMongoQueries bool
}

func (ref *QueryReflector) MakeReflector(val interface{}) (query Query, err error) {
	mongoVal, ok := val.(MongoQuery)
	if !ref.ReflectMongoQueries && ok {
		query = mongoVal
		return
	}

	v := reflect.ValueOf(val)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		err = errors.New("arcah: not struct")
		return
	}

	// desc, err := ref.DescriptorComputer.ComputeDescriptor(v.Type())
	// if err != nil {
	// 	return
	// }

	query = val
	return
}
