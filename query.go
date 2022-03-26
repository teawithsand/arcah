package arcah

type Query interface {
}

type keyValueQuery struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

type AndQuery []Query
type OrQuery []Query
type NotQuery struct {
	Query
}

type EqQuery keyValueQuery
type NeQuery keyValueQuery

type GtQuery keyValueQuery
type LtQuery keyValueQuery
type GteQuery keyValueQuery
type LteQuery keyValueQuery

// TODO(teawithsand): queries for arrays
