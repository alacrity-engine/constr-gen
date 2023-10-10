package main

import "go/ast"

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
