package acquery

import (
	"bytes"
	"encoding"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
)

type OrderField struct {
	Name string
	Desc bool
}

var _ encoding.TextMarshaler = &OrderField{}
var _ encoding.TextUnmarshaler = &OrderField{}

var ErrInvalidFirstChar = errors.New("arcah/acquery: OrderField first char must be '+' or '-' (either asc or desc ordering)")

func (f *OrderField) MarshalText() ([]byte, error) {
	res := make([]byte, len(f.Name)+1)
	if f.Desc {
		res[0] = '+'
	} else {
		res[0] = '-'
	}

	return res, nil
}

func (f *OrderField) UnmarshalText(data []byte) (err error) {
	if len(data) == 0 {
		err = ErrInvalidFirstChar
		return
	}

	if data[0] == '+' {
		f.Desc = false
	} else if data[0] == '-' {
		f.Desc = true
	} else {
		err = ErrInvalidFirstChar
		return
	}

	f.Name = string(data[1:])
	return
}

type OrderFields []OrderField

var _ encoding.TextMarshaler = &OrderFields{}
var _ encoding.TextUnmarshaler = &OrderFields{}

func (fields OrderFields) GetFields() bson.D {
	doc := bson.D{}
	for _, field := range fields {
		order := 1
		if field.Desc {
			order = -1
		}

		doc = append(doc, bson.E{
			Key:   field.Name,
			Value: order,
		})
	}
	return doc
}

func (fields *OrderFields) MarshalText() (res []byte, err error) {
	if fields == nil {
		return
	}
	for _, v := range *fields {
		var fieldRes []byte
		fieldRes, err = v.MarshalText()
		if err != nil {
			return
		}
		res = append(res, []byte(" ")...)
		res = append(res, fieldRes...)
	}

	return
}

func (fields *OrderFields) UnmarshalText(data []byte) (err error) {
	entries := bytes.Split(data, []byte(" "))
	for _, e := range entries {
		var f OrderField
		e = bytes.Trim(e, "\n\t\v ")
		err = f.UnmarshalText(e)
		if err != nil {
			return
		}
		*fields = append(*fields, f)
	}
	return
}
