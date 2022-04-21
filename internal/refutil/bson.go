package refutil

import (
	"strings"

	"github.com/teawithsand/reval/jsonutil"
)

const BsonTagName = "bson"

// Returns name of field, once it's rendered to BSON, when no such name is set by hand.
func DefaultBsonFieldName(name string) string {
	return strings.ToLower(name)
}

type BSONFieldMeta struct {
	BSONFieldName string
	Skip          bool
}

func (mtm *BSONFieldMeta) ParseTag(bsonTags string) (err error) {
	bsonFieldName, ok := jsonutil.GetJSONFieldName(bsonTags)
	if !ok {
		return
	}

	if len(bsonFieldName) == 0 {
		mtm.Skip = true
		return
	}

	mtm.BSONFieldName = bsonFieldName
	return
}
