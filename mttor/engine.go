package mttor

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/teawithsand/arcah/internal/refutil"
	"github.com/teawithsand/reval/stdesc"
	"go.mongodb.org/mongo-driver/bson"
)

// MutatorEngine is component responsible for applying mutations passed into it.
// Typically, there is single global mutator handling all types.
type MutatorEngine interface {
	// Applies specified mutation to target provided.
	Mutate(ctx context.Context, target, mutation interface{}) (err error)
}

// Mutator, which is able to apply mutations to mongo objects.
type MongoMutatorEngine interface {
	MutatorEngine
	RenderMongoMutation(ctx context.Context, targetType reflect.Type, mutation interface{}) (res interface{}, err error)
}

// Default implementation of Mutator, suitable for common tasks.
// It supports some most common tasks.
// It's also MongoMutator, with support for all mutations, which are MongoMutations.
type defaultMutatorEngine struct {
	mutationMap map[string]Mutator

	targetComputer   *stdesc.Computer
	mutationComputer *stdesc.Computer
}

func (dm *defaultMutatorEngine) Mutate(ctx context.Context, target, mutation interface{}) (err error) {
	refTarget := reflect.ValueOf(target)
	refMutation := reflect.ValueOf(mutation)

	targetDescriptor, err := dm.targetComputer.ComputeDescriptor(ctx, reflect.TypeOf(target))
	if err != nil {
		return
	}

	mutationDescriptor, err := dm.mutationComputer.ComputeDescriptor(ctx, reflect.TypeOf(mutation))
	if err != nil {
		return
	}

	for _, mf := range mutationDescriptor.NameToField {
		meta := mf.Meta.(mutatorMeta)
		// meta.TargetFieldName
		tf, ok := targetDescriptor.NameToField[meta.TargetFieldName]
		if !ok {
			err = &Error{
				Descriptorion: fmt.Sprintf("Field %s is not available in target of type %s", meta.TargetFieldName, refTarget.Type()),
			}
			return
		}

		mutation, ok := dm.mutationMap[meta.MutationName]
		if !ok {
			err = &Error{
				Descriptorion: fmt.Sprintf("Mutation %s is not registered", meta.MutationName),
			}
			return
		}

		mutationFieldRefValue := mf.MustGet(refMutation)

		if meta.TargetMutationArgs.IsSet("omitempty") {
			if refutil.ValueIsEmpty(mutationFieldRefValue) {
				continue
			}
		}

		err = mutation.ApplyMutation(ctx, refTarget, tf, MutatorData{
			Value:        mutationFieldRefValue.Interface(),
			Args:         meta.TargetMutationArgs,
			FieldName:    meta.TargetFieldName,
			MutationName: meta.MutationName,
		})
		if err != nil {
			return
		}
	}

	return
}

func (dm *defaultMutatorEngine) RenderMongoMutation(ctx context.Context, targetType reflect.Type, mutation interface{}) (res interface{}, err error) {
	refMutation := reflect.ValueOf(mutation)

	targetDescriptor, err := dm.targetComputer.ComputeDescriptor(ctx, targetType)
	if err != nil {
		return
	}

	mutationDescriptor, err := dm.mutationComputer.ComputeDescriptor(ctx, reflect.TypeOf(mutation))
	if err != nil {
		return
	}

	mutationRegistry := map[string]bson.D{}

	for _, mf := range mutationDescriptor.NameToField {
		meta := mf.Meta.(mutatorMeta)

		mutation, ok := dm.mutationMap[meta.MutationName]
		if !ok {
			err = &Error{
				Descriptorion: fmt.Sprintf("Mutation %s is not registered", meta.MutationName),
			}
			return
		}

		tf, ok := targetDescriptor.NameToField[meta.TargetFieldName]
		if !ok {
			err = &Error{
				Descriptorion: fmt.Sprintf("Field %s is not available in target of type %s", meta.TargetFieldName, targetType),
			}
			return
		}

		mongoMutation, ok := mutation.(MongoMutator)
		if !ok {
			err = &Error{
				Descriptorion: fmt.Sprintf("Registered mutation %s is not mongo mutation", meta.MutationName),
			}
			return
		}

		mfd := tf.Meta.(mutatorTargetMeta)

		if mfd.Skip {
			continue
		}

		mutationFieldRefValue := mf.MustGet(refMutation)

		if meta.TargetMutationArgs.IsSet("omitempty") {
			if refutil.ValueIsEmpty(mutationFieldRefValue) {
				continue
			}
		}

		mutationName := mongoMutation.MongoMutationName()
		_, ok = mutationRegistry[mutationName]
		if !ok {
			mutationRegistry[mutationName] = bson.D{}
		}

		var entry bson.E
		entry, err = mongoMutation.RenderMongoDoc(ctx, MongoMutatorData{
			MutatorData: MutatorData{
				Value:        mutationFieldRefValue.Interface(),
				Args:         meta.TargetMutationArgs,
				FieldName:    meta.TargetFieldName,
				MutationName: meta.MutationName,
			},
			BSONFieldName: mfd.BSONFieldName,
		})

		if err != nil {
			return
		}

		mutationRegistry[mutationName] = append(mutationRegistry[mutationName], entry)
	}

	innerRes := bson.D{}

	for k, v := range mutationRegistry {
		innerRes = append(innerRes, bson.E{
			Key:   k,
			Value: v,
		})
	}

	res = innerRes
	return
}

func NewDefaultMutatorEngine() (mutator MongoMutatorEngine) {
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
				err = meta.ParseTag(pf.Field.Tag.Get("bson"))
				if err != nil {
					return
				}

				hasNameSet := len(meta.BSONFieldName) > 0
				if !hasNameSet && !meta.Skip {
					meta.BSONFieldName = defaultBsonFieldName(pf.Field.Name)
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
