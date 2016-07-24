package main

import (
	"flag"
	"fmt"
	"github.com/aurelien-rainone/go-genstructs/lib"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
)

var (
	typename  = flag.String("type", "", "type name to inspect; must be set")
	generator = flag.String("gen", "", "generator function; must be set")
	filename  string
)

func showUsage() {

	usage := `go-genstructs - Automatically generate Go code from Go structures

USAGE:
   go-genstructs -type TYPE -gen FUNC FILE
   
VERSION:
   0.1.0
   
ARGUMENTS:

   FILE           file name where the structure; must be set if called
                  from command-line; automatically set if called from
                  'go generate'

GLOBAL OPTIONS:
   --type TYPE    type name of the structure to inspect; must be set
   --gen GENE     generator function name; must be set
   --help, -h     show help
   --version, -v  print the version
`
	fmt.Println(usage)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("go-genstructs: ")
	flag.Usage = showUsage
	flag.Parse()

	if *typename == "" || *generator == "" {
		flag.Usage()
		os.Exit(1)
	}

	file := ""
	if flag.NArg() > 0 {
		file = flag.Arg(0)
	} else {
		file = os.Getenv("GOFILE")
	}
	tdef, err := inspectCode(*typename, *generator, file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(tdef)
}

func inspectCode(typename, generator, filename string) (lib.Type, error) {
	var (
		typeDef    lib.Type // the Type struct we'll build and return
		inspecting bool     // are we actually inspecting the 'typename' ast?
	)

	fset := token.NewFileSet() // positions are relative to fset
	fileFilter := func(file os.FileInfo) bool {
		return file.Name() == filename
	}

	d, err := parser.ParseDir(fset, "./", fileFilter, parser.ParseComments)
	if err != nil {
		return typeDef, err
	}

	for _, f := range d {
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec:
				if x.Name.Name == typename {
					inspecting = true
					typeDef.Name = x.Name.Name
				} else if inspecting {
					// finished inspection
					inspecting = false
					return false
				}
			case *ast.Field:
				if inspecting {
					fmt.Println("len(Name):", len(x.Names))
					typeDef.Fields = append(typeDef.Fields, lib.Field{
						Name:    x.Names[0].Name,
						Comment: x.Doc.Text(),
					})
				}
			default:
				fmt.Println("other x:", x)
			}
			return true
		})
	}
	if typeDef.Name == "" {
		return typeDef, fmt.Errorf("type %s not found", typename)
	}
	if len(typeDef.Fields) == 0 {
		return typeDef, fmt.Errorf("type %s has 0 fields", typename)
	}
	return typeDef, nil
}
