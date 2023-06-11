// Command gen generates code to parse a CSV record to a struct.
package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"

	"github.com/nickng/csv"
)

var (
	typename = flag.String("type", "", "name of type to store a CSV; must be set")
	outfile  = flag.String("out", "parse_csv.generated.go", "filename of output file")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: gen -type typename\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Flags:")
	flag.PrintDefaults()
}

type Data struct {
	Package  string
	TypeName string
	Fields   []Field
}

type Field struct {
	CSVFieldName    string // CSV record field name
	StructFieldName string // Struct field name
}

//go:embed parse_csv.go.tmpl
var parseCSVTmpl string

func main() {
	log.SetFlags(0)
	log.SetPrefix("gen: ")
	flag.Parse()
	flag.Usage = Usage

	if *typename == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Tests: false,
	}
	pkgs, err := packages.Load(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	rowType, err := analyseType(*typename, pkgs[0])
	if err != nil {
		log.Fatal(err)
	}

	var d Data
	d.TypeName = *typename
	d.Package = pkgs[0].Name
	for _, field := range rowType.Fields.List {
		if field.Tag == nil {
			continue
		}
		if len(field.Names) != 1 {
			// Unknown field name format
			continue
		}
		csvTag := reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get("csv")
		tag := csv.ParseTag(csvTag)
		d.Fields = append(d.Fields, Field{
			CSVFieldName:    tag.FieldHeader,
			StructFieldName: field.Names[0].Name,
		})
	}

	tmpl := template.Must(template.New("").Parse(parseCSVTmpl))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		log.Fatal(err)
	}
	bs, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*outfile, bs, 0644); err != nil {
		log.Fatal(err)
	}
}

func analyseType(typename string, pkg *packages.Package) (*ast.StructType, error) {
	var (
		err        error
		structType *ast.StructType
	)
	for _, file := range pkg.Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			if err != nil {
				return false // Don't need to keep searching
			}
			decl, ok := node.(*ast.GenDecl)
			if !ok {
				return true
			}
			for _, spec := range decl.Specs {
				tspec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if tspec.Name.Name != typename {
					continue
				}
				t, ok := tspec.Type.(*ast.StructType)
				if !ok {
					err = fmt.Errorf("type %s is not a struct", typename)
					return false
				}
				structType = t
			}
			return false
		})
	}
	if err != nil {
		return nil, err
	}
	if structType == nil {
		return nil, fmt.Errorf("type %s not found", typename)
	}
	return structType, nil
}
