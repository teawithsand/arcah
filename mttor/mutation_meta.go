package mttor

import (
	"strings"

	"github.com/teawithsand/reval/jsonutil"
)

const defaultMutatorTagName = "mttor"

type MutationArgs map[string][]string

func (args MutationArgs) IsSet(arg string) bool {
	if args == nil {
		return false
	}

	_, ok := args[arg]
	return ok
}

func (args MutationArgs) GetFirst(arg string) string {
	if args == nil {
		return ""
	}

	v := args[arg]
	if len(v) > 0 {
		return v[0]
	}
	return ""
}

type mutatorMeta struct {
	// Name of field that this field mutates.
	// Defaults to field name from reflection.
	TargetFieldName string

	// Name of mutation to apply using value of this field and target.
	// By default set is used.
	MutationName string

	TargetMutationArgs MutationArgs
}

func (mm *mutatorMeta) ParseTag(tags string) (err error) {
	values := strings.Split(tags, ",")
	if len(values) >= 1 {
		mm.TargetFieldName = values[0]
	}
	if len(values) >= 2 {
		mm.MutationName = values[1]
	}

	args := MutationArgs{}

	if len(values) >= 3 {
		// TODO(teawithsand): enginem which allows
		for _, v := range values[2:] {
			res := strings.SplitN(v, ":", 1)
			if len(res) == 2 {
				args[res[0]] = append(args[res[0]], res[1])
			} else {
				args[res[0]] = append(args[res[0]], "")
			}
		}

	}
	mm.TargetMutationArgs = args

	return
}

type mutatorTargetMeta struct {
	BSONFieldName string
	Skip          bool
}

func (mtm *mutatorTargetMeta) ParseTag(bsonTags string) (err error) {
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
