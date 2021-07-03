package main

import (
	"go/ast"
	"io/fs"
)

// FileTracker cotains the info
// about the file and path to it.
type FileTracker struct {
	dir  string
	info fs.FileInfo
}

// TypeDeclaration contains the
// information about type declared
// in the game module.
type TypeDeclaration struct {
	packageName string
	typeName    string
	fieldList   []*ast.Field
}
