package acquery

import "context"

type MongoEngine interface {
	CompileQuery(ctx context.Context, query interface{}) (err error)
}
