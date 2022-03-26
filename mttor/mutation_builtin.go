package mttor

import (
	"context"
	"fmt"
	"reflect"

	"github.com/teawithsand/arcah/internal/refutil"
	"github.com/teawithsand/reval/stdesc"
	"go.mongodb.org/mongo-driver/bson"
)

type setMutation struct {
}

func (sm *setMutation) ApplyMutation(ctx context.Context, target reflect.Value, field stdesc.Field, data MutatorData) (err error) {
	field.MustSet(target, reflect.ValueOf(data.Value))
	return
}

func (sm *setMutation) MongoMutationName() string {
	return "$set"
}

func (sm *setMutation) RenderMongoDoc(ctx context.Context, data MongoMutatorData) (entry bson.E, err error) {
	return bson.E{
		Key:   data.BSONFieldName,
		Value: data.Value,
	}, nil
}

type incMutation struct {
}

func (sm *incMutation) ApplyMutation(ctx context.Context, target reflect.Value, field stdesc.Field, data MutatorData) (err error) {
	prevValue := refutil.ValueToNumber(field.MustGet(target))
	modValue := refutil.ValueToNumber(reflect.ValueOf(data.Value))

	if prevValue == nil {
		err = &Error{
			Descriptorion: fmt.Sprintf("inc mutation target field is not number"),
		}
		return
	}

	if modValue == nil {
		err = &Error{
			Descriptorion: fmt.Sprintf("mutator value target field is not number"),
		}
		return
	}

	if reflect.TypeOf(prevValue) != reflect.TypeOf(modValue) {
		err = &Error{
			Descriptorion: fmt.Sprintf("numbers in mutator and field have different types"),
		}
		return
	}

	var tempResult interface{}
	switch pv := prevValue.(type) {
	case int64:
		tempResult = pv + modValue.(int64)
	case uint64:
		tempResult = pv + modValue.(uint64)
	case float64:
		tempResult = pv + modValue.(float64)
	}

	field.MustSet(target, reflect.ValueOf(tempResult))
	return
}

func (sm *incMutation) MongoMutationName() string {
	return "$inc"
}

func (sm *incMutation) RenderMongoDoc(ctx context.Context, data MongoMutatorData) (entry bson.E, err error) {
	return bson.E{
		Key:   data.BSONFieldName,
		Value: data.Value,
	}, nil
}

type pushMutation struct {
}

func (sm *pushMutation) ApplyMutation(ctx context.Context, target reflect.Value, field stdesc.Field, data MutatorData) (err error) {
	targetFieldValue := field.MustGet(target)
	if targetFieldValue.Type().Kind() != reflect.Slice {
		err = &Error{
			Descriptorion: fmt.Sprintf("target is not slice"),
		}
		return
	}

	pushValue := reflect.ValueOf(data.Value)
	if pushValue.Type() == targetFieldValue.Type().Elem() {
		targetFieldValue = reflect.Append(targetFieldValue, pushValue)
	} else if pushValue.Type().Kind() == reflect.Slice && pushValue.Type().Elem() == targetFieldValue.Type().Elem() {
		targetFieldValue = reflect.AppendSlice(targetFieldValue, pushValue)
	} else if pushValue.Type().Kind() == reflect.Array && pushValue.Type().Elem() == targetFieldValue.Type().Elem() {
		targetFieldValue = reflect.AppendSlice(targetFieldValue, pushValue.Slice(0, pushValue.Len()))
	} else {
		err = &Error{
			Descriptorion: fmt.Sprintf("element of type %T is not compatible with slice of type %T", data.Value, targetFieldValue.Interface()),
		}
		return
	}

	field.MustSet(target, targetFieldValue)
	return
}

func (sm *pushMutation) MongoMutationName() string {
	return "$push"
}

func (sm *pushMutation) RenderMongoDoc(ctx context.Context, data MongoMutatorData) (entry bson.E, err error) {
	var value interface{}
	if reflect.TypeOf(data.Value).Kind() != reflect.Slice {
		sliceValue := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(data.Value)), 1, 1)
		sliceValue.Index(0).Set(reflect.ValueOf(data.Value))
		value = sliceValue.Interface()
	} else {
		value = data.Value
	}
	return bson.E{
		Key: data.BSONFieldName,
		Value: bson.D{
			bson.E{
				Key:   "$each",
				Value: value,
			},
		},
	}, nil
}
