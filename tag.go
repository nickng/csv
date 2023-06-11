package csv

import (
	"strings"
)

// Tag represents a "csv" struct field tag.
//
// For example, `csv:"field_name"` is represented as Tag{FieldName: "field_name"}
type Tag struct {
	// FieldHeader is the CSV header value of the field.
	FieldHeader string
	Options     string
}

// ParseTag parses a raw struct tag (`csv:"tag,value,value2"`)
// and returns a Tag representing its content.
func ParseTag(tag string) Tag {
	name, opts, _ := strings.Cut(tag, ",")
	return Tag{FieldHeader: name, Options: opts}
}
