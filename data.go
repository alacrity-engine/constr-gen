package main

import (
	"os"
)

// FileTracker cotains the info
// about the file and path to it.
type FileTracker struct {
	dir  string
	info os.DirEntry
}

// ComponentData contains the
// information about type declared
// in the game module.
type ComponentData struct {
	Imports map[string]struct{}
	Name    string
	PkgPath string
	Fields  []FieldData
}

type FieldData struct {
	Name    string
	IsArray bool
	IsMap   bool
	Type    string
}
