package mttor

import "strings"

// Returns name of field, once it's rendered to BSON, when no such name is set by hand.
func defaultBsonFieldName(name string) string {
	return strings.ToLower(name)
}
