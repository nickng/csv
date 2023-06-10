package csv

import (
	"fmt"
	"strings"
)

// Tag represents a "csv" struct field tag.
//
// For example, `csv:"field_name"` is represented as Tag{FieldName: "field_name"}
type Tag struct {
	// FieldHeader is the CSV header value of the field.
	FieldHeader string
}

// ParseTag parses a raw struct tag (`csv:"tag,value,value2"`)
// and returns a Tag representing its content.
func ParseTag(tag string) Tag {
	fullTag := tag
	if !strings.ContainsRune(tag, ',') {
		fullTag = fmt.Sprintf("%s,", tag)
	}
	parts := strings.Split(fullTag, ",")
	if len(parts) > 1 && parts[0] != "-" {
		return Tag{FieldHeader: parts[0]}
	}
	return Tag{}
}
