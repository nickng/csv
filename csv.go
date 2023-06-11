package csv

import (
	"encoding/csv"
	"fmt"
	"reflect"
)

// Reader is a structured data reader from CSV.
type Reader[T any] struct {
	rd           *csv.Reader // Underlying CSV reader
	fieldIndex   map[int]int // Converts record field index to struct field index
	parsedHeader bool
}

// NewReader creates a new structured data reader from an underlying
// raw CSV record reader. It returns error if the generic type T is
// not a valid type to stored the parsed data.
func NewReader[T any](r *csv.Reader) (*Reader[T], error) {
	csvReader := &Reader[T]{rd: r}
	if err := csvReader.validateFields(); err != nil {
		return nil, err
	}
	return csvReader, nil
}

var (
	errNotPointer         = fmt.Errorf("fields should be a pointer")
	errNotStructPointer   = fmt.Errorf("fields should be a pointer to a struct")
	errFieldNotAssignable = fmt.Errorf("field is not assignable")
)

// validateFieldsType checks that the generic type T can be used to store
// record field values of a CSV file. T should be a pointer to a struct.
// All tagged fields should be string.
func (r *Reader[T]) validateFields() error {
	var rowPtr T
	rowPtrType := reflect.TypeOf(rowPtr) // reflect.Value of rowPtr
	if rowPtrType.Kind() != reflect.Pointer {
		return errNotPointer
	}
	rowStruct := rowPtrType.Elem()
	if rowStruct.Kind() != reflect.Struct {
		return errNotStructPointer
	}
	for i := 0; i < rowStruct.NumField(); i++ {
		f := rowStruct.Field(i)
		tag := ParseTag(f.Tag.Get("csv"))
		if tag.FieldHeader != "" {
			if rowStruct.FieldByIndex([]int{i}).Type.Kind() != reflect.String {
				return fmt.Errorf("invalid field %s: %w", rowStruct.Field(i).Name, errFieldNotAssignable)
			}
		}
	}
	return nil
}

// parseHeader parses the header row of the CSV and prepares to store
// record fields to variables of type T.
func (r *Reader[T]) parseHeader(header []string, rowPtr T) error {
	headerToIndex := make(map[string]int)
	for i, field := range header {
		headerToIndex[field] = i
	}
	rowStruct := reflect.Indirect(reflect.ValueOf(rowPtr))
	for i := 0; i < rowStruct.NumField(); i++ {
		f := rowStruct.Type().Field(i)
		tag := ParseTag(f.Tag.Get("csv"))
		if r.fieldIndex == nil {
			r.fieldIndex = make(map[int]int)
		}
		if _, exists := headerToIndex[tag.FieldHeader]; !exists {
			// Tag specifies a field that isn't in the header, all
			// records will use zero value for that struct field.
			continue
		}
		r.fieldIndex[headerToIndex[tag.FieldHeader]] = i
	}
	return nil
}

// assignFields takes a record and assigns to rowPtr struct.
func (r *Reader[T]) assignFields(record []string, rowPtr T) error {
	for i, field := range record {
		sfIndex, exists := r.fieldIndex[i]
		if !exists {
			continue
		}
		rowStruct := reflect.Indirect(reflect.ValueOf(rowPtr))
		rowStruct.FieldByIndex([]int{sfIndex}).SetString(field)
	}
	return nil
}

// Read reads one record as rowPtr.
// It returns io.EOF if there's no more record to read.
func (r *Reader[T]) Read(rowPtr T) error {
	if !r.parsedHeader {
		rcd, err := r.rd.Read()
		if err != nil {
			return err
		}
		if err := r.parseHeader(rcd, rowPtr); err != nil {
			return err
		}
		r.parsedHeader = true
	}
	rcd, err := r.rd.Read()
	if err != nil {
		return err
	}
	if err := r.assignFields(rcd, rowPtr); err != nil {
		return err
	}
	return nil
}
