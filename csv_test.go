package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"testing"

	_ "embed"
)

//go:embed testdata/example.csv
var exampleCSV string

// exampleType corresponds to exampleCSV
type exampleType struct {
	Bar string `csv:"bar"`
	Baz string `csv:"baz"`
	Foo string `csv:"foo"`
}

func TestReader_validateFieldsType(t *testing.T) {
	// not assignable
	r := &Reader[exampleType]{}
	if want, got := errNotPointer, r.validateFields(); !errors.Is(got, want) {
		t.Fatalf("expected error %v but got %v", want, got)
	}
	// wrong type
	r2 := &Reader[*struct {
		Field int `csv:"field"`
	}]{}
	if want, got := errFieldNotAssignable, r2.validateFields(); !errors.Is(got, want) {
		t.Fatalf("expected error %v but got %v", want, got)
	}
}

func TestReader_parseHeader(t *testing.T) {
	r, err := NewReader[*exampleType](csv.NewReader(strings.NewReader(exampleCSV)))
	if err != nil {
		t.Fatalf("expected no error for creating reader but got %v", err)
	}
	// Use the underlying CSV reader to read the header row
	header, err := r.rd.Read()
	if err != nil {
		t.Fatalf("expected no error for reading header but got %v", err)
	}
	var record exampleType
	if err := r.parseHeader(header, &record); err != nil {
		t.Fatalf("expected no error for parsing header but got %v", err)
	}

	// The order of the header in the file is foo,bar,baz
	// and in exampleType the struct field order is Bar Baz Foo
	//
	// So the mapping between header field order is 0→2, 1→0, 2→1
	testCases := [...]struct {
		headerIndex, structFieldIndex int
	}{
		{headerIndex: 0, structFieldIndex: 2},
		{headerIndex: 1, structFieldIndex: 0},
		{headerIndex: 2, structFieldIndex: 1},
	}
	for _, tc := range testCases {
		if want, got := tc.structFieldIndex, r.fieldIndex[tc.headerIndex]; want != got {
			t.Fatalf("expected header index %d → struct field index %d but got struct field index %d", tc.headerIndex, want, got)
		}
	}
}

func TestReader_assignFields(t *testing.T) {
	// CSV header: foo,bar,baz
	// Struct fields: Bar Baz Foo
	r := &Reader[*exampleType]{fieldIndex: map[int]int{0: 2, 1: 0, 2: 1}}

	testCases := [...]struct {
		name           string
		record         []string
		expectedRecord exampleType
	}{
		{name: "happy path", record: []string{"1", "2", "hello"}, expectedRecord: exampleType{Foo: "1", Bar: "2", Baz: "hello"}},
		{name: "more fields in data than expected", record: []string{"1", "2", "hello", "world (ignored)"}, expectedRecord: exampleType{Foo: "1", Bar: "2", Baz: "hello"}},
	}

	for _, tc := range testCases {
		var record exampleType
		if err := r.assignFields(tc.record, &record); err != nil {
			t.Fatalf("expected no error but got %v", err)
		}
		if want, got := tc.expectedRecord, record; want != got {
			t.Fatalf("expected record to be %v but got %v", want, got)
		}
	}
}

func TestReader(t *testing.T) {
	r, err := NewReader[*exampleType](csv.NewReader(strings.NewReader(exampleCSV)))
	if err != nil {
		t.Fatalf("expected no error for creating reader but got %v", err)
	}
	var record exampleType
	if err := r.Read(&record); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if want, got := (exampleType{Foo: "1", Bar: "2", Baz: "hello"}), record; want != got {
		t.Fatalf("expecting %v but got %v", want, got)
	}
	var record2 exampleType
	if err := r.Read(&record2); err != nil {
		t.Fatalf("expected no error but got %v", err)
	}
	if want, got := (exampleType{Foo: "3", Bar: "2", Baz: "world"}), record2; want != got {
		t.Fatalf("expecting %v but got %v", want, got)
	}
	var record3 exampleType
	if err := r.Read(&record3); err != io.EOF {
		t.Fatalf("expected EOF error but got none")
	}
}

// If a file has no header, Reader will return the underlying io.EOF error.
func TestReader_emptyFile(t *testing.T) {
	emptyFile := ``
	r, err := NewReader[*exampleType](csv.NewReader(strings.NewReader(emptyFile)))
	if err != nil {
		t.Fatalf("expected no error for creating reader but got %v", err)
	}
	var record exampleType
	if err := r.Read(&record); err != io.EOF {
		t.Fatalf("expected EOF error but got none")
	}
}

// If a file has only header, it's equivalent to reading
// an empty file with Reader because Reader will process
// the header then try to read the content.
func TestReader_headerOnly(t *testing.T) {
	headerOnly := "foo,bar,baz\n"
	r, err := NewReader[*exampleType](csv.NewReader(strings.NewReader(headerOnly)))
	if err != nil {
		t.Fatalf("expected no error for creating reader but got %v", err)
	}
	var record exampleType
	if err := r.Read(&record); err != io.EOF {
		t.Fatalf("expected EOF error but got none")
	}
}

func ExampleReader() {
	r, err := NewReader[*exampleType](csv.NewReader(strings.NewReader(exampleCSV)))
	if err != nil {
		log.Fatal(err)
	}
	for {
		var record exampleType
		if err := r.Read(&record); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", record)
	}
	// Output:
	// {Bar:2 Baz:hello Foo:1}
	// {Bar:2 Baz:world Foo:3}
}
