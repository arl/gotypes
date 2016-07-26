package gotypes

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

type Field struct {
	Type string            // field type
	Name string            // field name
	Doc  string            // field documentation
	Meta map[string]string // client-added meta data
}

type Struct struct {
	Name   string
	Fields []Field
}

func (s *Struct) addField(t, n, d string) {
	s.Fields = append(s.Fields,
		Field{
			Type: t,
			Name: n,
			Doc:  d,
			Meta: make(map[string]string),
		})
}

func Inspect(typename, filename string) (Struct, error) {
	var (
		structDef  Struct // the Struct we'll build and return
		inspecting bool   // are we actually inspecting the 'typename' ast?
	)

	fset := token.NewFileSet() // positions are relative to fset
	fileFilter := func(file os.FileInfo) bool {
		return file.Name() == filename
	}

	d, err := parser.ParseDir(fset, "./", fileFilter, parser.ParseComments)
	if err != nil {
		return structDef, err
	}

	for _, f := range d {
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				if x.Name.Name == typename {
					inspecting = true
					structDef.Name = x.Name.Name
				} else if inspecting {
					// finished inspection
					inspecting = false
					return false
				}
			case *ast.Field:
				if inspecting {
					if len(x.Names) > 1 {
						fmt.Println("len(Name):", len(x.Names))
						log.Fatalf("Can't have this kind of declaration")
					}
					ident := x.Type.(*ast.Ident)
					structDef.addField(ident.Name, x.Names[0].Name,
						strings.TrimSpace(x.Doc.Text()))
				}
			default:
			}
			return true
		})
	}
	if structDef.Name == "" {
		return structDef, fmt.Errorf("type %s not found", typename)
	}
	if len(structDef.Fields) == 0 {
		return structDef, fmt.Errorf("type %s has 0 fields", typename)
	}
	return structDef, nil
}
