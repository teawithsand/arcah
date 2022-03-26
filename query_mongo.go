package arcah

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

type MongoQuery interface {
	RenderMongo() (res interface{}, err error)
}

type MongoQueryRenderer struct {
}

func mongoRenderKeyValueQuery(op string, query keyValueQuery) interface{} {
	return bson.D{
		{op, bson.D{
			{query.Field, query.Value},
		}},
	}
}

func (mqr *MongoQueryRenderer) innerRender(query Query) (res interface{}, err error) {
	switch typedQuery := query.(type) {
	case MongoQuery:
		return typedQuery.RenderMongo()
	case AndQuery:
		mapped := make([]interface{}, 0, len(typedQuery))
		for _, q := range typedQuery {
			var rendered interface{}
			rendered, err = mqr.Render(q)
			if err != nil {
				return
			}

			if rendered != nil {
				mapped = append(mapped, rendered)
			}
		}

		res = bson.M{
			"$and": mapped,
		}
		return
	case OrQuery:
		mapped := make([]interface{}, 0, len(typedQuery))
		for _, q := range typedQuery {
			var rendered interface{}
			rendered, err = mqr.Render(q)
			if err != nil {
				return
			}

			if rendered != nil {
				mapped = append(mapped, rendered)
			}
		}

		res = bson.M{
			"$or": mapped,
		}
		return
	case NotQuery:
		res = bson.M{
			"$not": typedQuery.Query,
		}
		return
	case EqQuery:
		res = mongoRenderKeyValueQuery("$eq", keyValueQuery(typedQuery))
		return
	case NeQuery:
		res = mongoRenderKeyValueQuery("$ne", keyValueQuery(typedQuery))
		return
	case LtQuery:
		res = mongoRenderKeyValueQuery("$lt", keyValueQuery(typedQuery))
		return
	case LteQuery:
		res = mongoRenderKeyValueQuery("$lte", keyValueQuery(typedQuery))
		return
	case GtQuery:
		res = mongoRenderKeyValueQuery("$gt", keyValueQuery(typedQuery))
		return
	case GteQuery:
		res = mongoRenderKeyValueQuery("$gte", keyValueQuery(typedQuery))
		return
	default:
		err = fmt.Errorf("query type %T is not supported", typedQuery)
		return
	}
}

func (mqr *MongoQueryRenderer) Render(query Query) (res interface{}, err error) {
	return mqr.innerRender(query)
}
