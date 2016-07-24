package main

import (
	"fmt"
	"github.com/urfave/cli"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

type Field struct {
	Name    string
	Comment string
}

type Type struct {
	Name   string
	Fields []Field
}

func inspectType(typename, filename string) (Type, error) {
	var (
		typeDef    Type // the Type struct we'll build and return
		inspecting bool // are we actually inspecting the 'typename' ast?
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
					typeDef.Fields = append(typeDef.Fields, Field{
						Name:    x.Names[0].Name,
						Comment: x.Doc.Text(),
					})
				}
			default:
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

func main() {

	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "type",
			Value: "",
			Usage: "type name to inspect; must be set",
		},
	}

	app.Action = func(c *cli.Context) error {
		file := ""
		if c.NArg() > 0 {
			file = c.Args().Get(0)
		} else {
			file = os.Getenv("GOFILE")
		}
		tdef, err := inspectType(c.String("type"), file)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(tdef)
		return err
	}

	app.Run(os.Args)
}
