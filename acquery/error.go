package acquery

import "errors"

var ErrInvalidOrderFields = errors.New("arcah/acquery: invalid order fields provided")
var ErrFieldTwice = errors.New("arcah/acquery: some field was found twice")
