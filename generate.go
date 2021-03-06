package gotypes

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"
	"text/template"
)

// Generator holds the state of the analysis. Primarily used to buffer
// the output for format.Source.
type Generator struct {
	buf        bytes.Buffer // Accumulated output.
	tmpl       string       // template string
	structData Struct       // template input struct
}

func (g *Generator) Printf(format string, args ...interface{}) {
	fmt.Fprintf(&g.buf, format, args...)
}

// format returns the gofmt-ed contents of the Generator's buffer.
func (g *Generator) format(dbg bool) []byte {
	// create container template
	var (
		tmpl *template.Template
		err  error
	)
	if tmpl, err = template.New(os.Args[0]).Parse(g.tmpl); err != nil {
		panic(err)
	}
	err = tmpl.Execute(&g.buf, g.structData)
	if err != nil {
		panic(err)
	}

	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		if !dbg {
			log.Printf("warning: compile the package to analyze the error")
		}
		return g.buf.Bytes()
	}
	if dbg {
		log.Println("Generated source:\n", string(src))
	}
	return src
}

func Generate(template string, structData Struct, dbg bool) []byte {
	// create the generator
	g := Generator{
		tmpl:       template,
		structData: structData,
	}

	// Print the header and package clause.
	g.Printf("// This file has been generated by \"%s\"; DO NOT EDIT\n", os.Args[0])
	g.Printf("// command: \"%s\"\n", strings.Join(os.Args, " "))
	g.Printf("\n")

	return g.format(dbg)
}
