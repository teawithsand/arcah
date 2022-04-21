package acquery

var NoLimitPagination = Pagination{
	Offset: 0,
	Limit:  ^uint32(0),
}

// Base type of pagination used in acquery.
type Pagination struct {
	Offset uint32 `json:"offset" schema:"offset"`
	Limit  uint32 `json:"limit" schema:"limit"`
}

func min(x, y uint32) uint32 {
	if x < y {
		return x
	} else {
		return y
	}
}

func (p *Pagination) WithMaxLimit(limit uint32) Pagination {
	pp := *p
	pp.Limit = min(pp.Limit, limit)
	return pp
}
