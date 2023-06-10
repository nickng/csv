package csv

import "testing"

func TestParseTag(t *testing.T) {
	tcs := [...]struct {
		name        string
		tag         string
		expectedTag Tag
	}{
		{name: "empty", tag: "", expectedTag: Tag{}},
		{name: "just header", tag: "field_header", expectedTag: Tag{FieldHeader: "field_header"}},
		{name: "empty header", tag: ",", expectedTag: Tag{FieldHeader: ""}},
		{name: "explicit ignore header", tag: "-,", expectedTag: Tag{FieldHeader: ""}},
		{name: "explicit ignore header without trailing comma", tag: "-", expectedTag: Tag{}},
		{name: "header with other unrecognised options", tag: "field_header,omitempty,irrelevant", expectedTag: Tag{FieldHeader: "field_header"}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(tag string, expectedTag Tag) func(*testing.T) {
			return func(t *testing.T) {
				parsed := ParseTag(tag)
				if want, got := expectedTag, parsed; want != got {
					t.Fatalf("expected tag `%s` to be parsed as %+v but got %+v", tag, want, got)
				}
			}
		}(tc.tag, tc.expectedTag))
	}
}
