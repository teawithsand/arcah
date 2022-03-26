package mttor

import (
	"context"
	"reflect"

	"github.com/teawithsand/reval/stdesc"
	"go.mongodb.org/mongo-driver/bson"
)

type MutatorData struct {
	Value interface{}

	Args         MutationArgs
	FieldName    string
	MutationName string
}

// Muatator is part of DefaultMutator, which applies mutation using data it's given.
type Muatator interface {
	ApplyMutation(ctx context.Context, target reflect.Value, field stdesc.Field, data MutatorData) (err error)
}

type MongoMutatorData struct {
	MutatorData
	BSONFieldName string
	Skip          bool
}

// Mutation, which is able to render itself as mongodb mutation.
type MongoMutator interface {
	Muatator
	// Returns mongo mutation name, so appropriate bson.D can be passed.
	MongoMutationName() string
	RenderMongoDoc(ctx context.Context, data MongoMutatorData) (entry bson.E, err error)
}
