package main

import (
	"go/ast"
	"strings"
)

func fieldDefinition(field *ast.Field) (string, string) {
	// Assume the type of the field.
	var paramTypeName string

	switch t := field.Type.(type) {
	// If it's a simple identifier (such as int, float64, etc.).
	case *ast.Ident:
		paramTypeName = t.Name

	// If it's an array.
	case *ast.ArrayType:
		switch tt := t.Elt.(type) {
		case *ast.Ident:
			paramTypeName = "[]" + tt.Name

		case *ast.SelectorExpr:
			x, ok := tt.X.(*ast.Ident)

			if !ok {
				return "", ""
			}

			paramTypeName = "[]" + x.Name +
				"." + tt.Sel.Name

		default:
			return "", ""
		}

	// If it'a map (aka key-value dictionary).
	case *ast.MapType:
		// Assume the type of map key.
		var keyType string

		switch tt := t.Key.(type) {
		case *ast.Ident:
			keyType = tt.Name

		case *ast.SelectorExpr:
			x, ok := tt.X.(*ast.Ident)

			if !ok {
				return "", ""
			}

			keyType = x.Name + "." + tt.Sel.Name
		}

		// Assume the type of map value.
		var valueType string

		switch tt := t.Value.(type) {
		case *ast.Ident:
			valueType = tt.Name

		case *ast.SelectorExpr:
			x, ok := tt.X.(*ast.Ident)

			if !ok {
				return "", ""
			}

			valueType = x.Name + "." + tt.Sel.Name
		}

		paramTypeName = "map[" + keyType + "]" + valueType

	// If it's a structure from some package
	// (such as time.Duration or pixel.Vec).
	case *ast.SelectorExpr:
		x, ok := t.X.(*ast.Ident)

		if !ok {
			return "", ""
		}

		paramTypeName = x.Name + "." + t.Sel.Name

	// If it's an object.
	case *ast.StarExpr:
		switch inner := t.X.(type) {
		// If it's an internal object.
		case *ast.Ident:
			paramTypeName = "*" + inner.Name

		// If it's an external object.
		case *ast.SelectorExpr:
			x, ok := inner.X.(*ast.Ident)

			if !ok {
				return "", ""
			}

			paramTypeName = "*" + x.Name + "." + inner.Sel.Name
		}

	default:
		return "", ""
	}

	paramName := field.Names[0].Name

	return paramName, paramTypeName
}

func parseFieldList(fieldList []*ast.Field) []FieldData {
	fields := make([]FieldData, 0, len(fieldList))

	for _, field := range fieldList {
		// If the field is unnamed, skip it.
		if len(field.Names) <= 0 {
			continue
		}

		// If the field is not public, skip it.
		if field.Tag == nil || !strings.Contains(field.Tag.Value, `iris:"exported"`) {
			continue
		}

		paramName, paramTypeName := fieldDefinition(field)

		if paramName == "" || paramTypeName == "" {
			continue
		}

		field := FieldData{
			Name: paramName,
			Type: paramTypeName,
		}
		fields = append(fields, field)
	}

	return fields
}

func isComponent(fields []*ast.Field) bool {
	found := false

	// Find the "engine.BaseComponent".
	for _, field := range fields {
		t, ok := field.Type.(*ast.SelectorExpr)

		if !ok {
			continue
		}

		x, ok := t.X.(*ast.Ident)

		if !ok {
			continue
		}

		if t.Sel.Name == "BaseComponent" && x.Name == "engine" {
			found = true
			break
		}
	}

	return found
}

func findComponents(file *ast.File) ([]ComponentData, error) {
	typeDecls := []ComponentData{}

	ast.Inspect(file, func(n ast.Node) bool {
		switch t := n.(type) {
		// Find the type declaration.
		case *ast.GenDecl:
			for _, spec := range t.Specs {
				switch tt := spec.(type) {
				case *ast.TypeSpec:
					// Read the type name.
					typeName := tt.Name.Name

					switch ttt := tt.Type.(type) {
					case *ast.StructType:
						// Read the list of type fields.
						fieldList := ttt.Fields.List

						if !isComponent(fieldList) {
							continue
						}

						typeDecl := ComponentData{
							Name:   typeName,
							Fields: parseFieldList(fieldList),
						}

						typeDecls = append(typeDecls, typeDecl)
					}
				}
			}
		}

		return true
	})

	return typeDecls, nil
}

func getImportSet(file *ast.File) map[string]struct{} {
	imports := map[string]struct{}{}

	for _, imp := range file.Imports {
		importPath := imp.Path.Value

		if imp.Name != nil {
			importPath = imp.Name.Name + " " + importPath
		}

		imports[importPath] = struct{}{}
	}

	return imports
}
