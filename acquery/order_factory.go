package acquery

import (
	"context"
	"reflect"
	"sync"

	"github.com/teawithsand/arcah/internal/refutil"
	"github.com/teawithsand/reval/stdesc"
)

const orderTagName = "order"

type OrderSchemaFactory interface {
	CreateOrderSchema(ctx context.Context, ty reflect.Type) (schema *OrderSchema, err error)
}

type mongoOrderSchemaFactory struct {
	computer *stdesc.Computer
}

func NewMongoOrderSchemaFactory() OrderSchemaFactory {
	return &mongoOrderSchemaFactory{
		computer: &stdesc.Computer{
			FieldProcessorFactory: stdesc.FieldProcessorFunc(func(pf stdesc.PendingFiled) (options stdesc.FieldOptions, err error) {
				meta := refutil.BSONFieldMeta{}
				err = meta.ParseTag(pf.Field.Tag.Get(refutil.BsonTagName))
				if err != nil {
					return
				}

				options.Name = meta.BSONFieldName
				options.Skip = meta.Skip
				// TODO(teawithsand): support for embedding
				// options.Embed =

				return
			}),
			Summarizer: stdesc.SummarizerFunc(func(ctx context.Context, desc stdesc.Descriptor) (meta interface{}, err error) {
				innerSchema := &OrderSchema{}

				for name, f := range desc.NameToField {
					innerSchema = innerSchema.AddField(name, f.Name)
				}

				meta = innerSchema
				return
			}),

			Cache: &sync.Map{},
		},
	}
}

var _ OrderSchemaFactory = &mongoOrderSchemaFactory{}

// Creates order schema for mongodb queries.
// Uses field names from metadata provided or field name as aliases.
// Uses BSON names as db names.
func (osf *mongoOrderSchemaFactory) CreateOrderSchema(ctx context.Context, ty reflect.Type) (schema *OrderSchema, err error) {
	desc, err := osf.computer.ComputeDescriptor(ctx, ty)
	if err != nil {
		return
	}

	schema = desc.ComputedSummary.(*OrderSchema)
	return
}
