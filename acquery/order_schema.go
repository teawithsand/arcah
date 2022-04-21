package acquery

type OrderSchema struct {
	AliasToField map[string]string
}

func (schema *OrderSchema) AddField(userAlias, dbName string) *OrderSchema {
	if schema.AliasToField == nil {
		schema.AliasToField = make(map[string]string)
	}

	if len(dbName) == 0 {
		dbName = userAlias
	}

	schema.AliasToField[userAlias] = dbName

	return schema
}

func (schema *OrderSchema) Validate(fields OrderFields) (err error) {
	if len(schema.AliasToField) == 0 {
		if len(fields) != 0 {
			err = ErrInvalidOrderFields
			return
		}
		return
	}
	usedMap := make(map[string]struct{})

	for _, f := range fields {
		_, ok := usedMap[f.Name]
		if ok {
			err = ErrFieldTwice
			return
		}

		_, ok = schema.AliasToField[f.Name]
		if !ok {
			err = ErrInvalidOrderFields
			return
		}

		usedMap[f.Name] = struct{}{}
	}

	return
}

// Processes fields, and applies aliases
func (schema *OrderSchema) Process(fields OrderFields) (res OrderFields) {
	res = make(OrderFields, 0, len(fields)/2)
	if schema.AliasToField == nil {
		return
	}
	usedMap := make(map[string]struct{})
	for _, aliasField := range fields {
		_, ok := usedMap[aliasField.Name]
		if ok {
			continue
		}

		dbName, ok := schema.AliasToField[aliasField.Name]
		if ok {
			res = append(res, OrderField{
				Name: dbName,
				Desc: aliasField.Desc,
			})
		}

		usedMap[aliasField.Name] = struct{}{}
	}
	return
}
