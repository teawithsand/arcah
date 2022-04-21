package mttor

import (
	"context"
	"fmt"
	"reflect"

	"github.com/teawithsand/arcah/internal/refutil"
	"github.com/teawithsand/reval/stdesc"
	"go.mongodb.org/mongo-driver/bson"
)

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
