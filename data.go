package main

import (
	"go/ast"
	"os"
)

// FileTracker cotains the info
// about the file and path to it.
type FileTracker struct {
	dir  string
	info os.DirEntry
}

// TypeDeclaration contains the
// information about type declared
// in the game module.
type TypeDeclaration struct {
	packageName string
	packagePath string
	typeName    string
	fieldList   []*ast.Field
	imports     map[string]struct{}
}

type ComponentTemplateData struct {
	Imports map[string]struct{}
	Name    string
	PkgPath string
	Fields  []ComponentFieldTemplateData
}

type ComponentFieldTemplateData struct {
	Name string
	Type string
}
