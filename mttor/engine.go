package mttor

import (
	"context"
	"reflect"
	"sync"

	"github.com/teawithsand/arcah/internal/refutil"
	"github.com/teawithsand/reval/stdesc"
)

// Engine is component responsible for applying mutations passed into it.
// Typically, there is single global mutator handling all types.
type Engine interface {
	// Applies specified mutation to target provided.
	Mutate(ctx context.Context, target, mutation interface{}) (err error)
}

// Mutator, which is able to apply mutations to mongo objects.
type MongoEngine interface {
	RenderMongoMutation(ctx context.Context, targetType reflect.Type, mutation interface{}) (res interface{}, err error)
}

func NewMongoEngine() (mutator MongoEngine) {
	return NewDefaultEngine().(MongoEngine)
}

func NewDefaultEngine() (mutator Engine) {
	mutator = &defaultMutatorEngine{
		mutationMap: map[string]Mutator{
			"":     &setMutation{}, // default is set
			"set":  &setMutation{},
			"inc":  &incMutation{},
			"push": &pushMutation{},
		},
		targetComputer: &stdesc.Computer{
			Cache: &sync.Map{},
			FieldProcessorFactory: stdesc.FieldProcessorFunc(func(pf stdesc.PendingFiled) (options stdesc.FieldOptions, err error) {
				var meta mutatorTargetMeta
				err = meta.ParseTag(pf.Field.Tag.Get(refutil.BsonTagName))
				if err != nil {
					return
				}

				hasNameSet := len(meta.BSONFieldName) > 0
				if !hasNameSet && !meta.Skip {
					meta.BSONFieldName = refutil.DefaultBsonFieldName(pf.Field.Name)
				}

				options.Name = pf.Field.Name
				options.Meta = meta

				options.Embed = (pf.Field.Anonymous && pf.Field.Type.Kind() == reflect.Struct ||
					(pf.Field.Type.Kind() == reflect.Ptr && pf.Field.Type.Elem().Kind() == reflect.Struct)) &&
					!hasNameSet && !meta.Skip
				return
			}),
		},
		mutationComputer: &stdesc.Computer{
			FieldProcessorFactory: stdesc.FieldProcessorFunc(func(pf stdesc.PendingFiled) (options stdesc.FieldOptions, err error) {
				var meta mutatorMeta

				// ignore err from parsing tags for now
				err = meta.ParseTag(pf.Field.Tag.Get(defaultMutatorTagName))
				if err != nil {
					return
				}

				options.Skip = !pf.Field.IsExported() || meta.TargetFieldName == "-"

				if len(meta.TargetFieldName) == 0 {
					meta.TargetFieldName = pf.Field.Name
				}

				options.Name = pf.Field.Name
				options.Meta = meta
				options.Embed = (pf.Field.Anonymous && pf.Field.Type.Kind() == reflect.Struct ||
					(pf.Field.Type.Kind() == reflect.Ptr && pf.Field.Type.Elem().Kind() == reflect.Struct)) && meta.MutationName == ""
				return
			}),
			Cache: &sync.Map{},
		},
	}
	return
}
