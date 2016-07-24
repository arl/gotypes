package lib

type Field struct {
	Name    string
	Comment string
}

type Type struct {
	Name   string
	Fields []Field
}
