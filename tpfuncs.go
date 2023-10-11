package main

import (
	"strings"
	"text/template"
)

var templateFuncs = template.FuncMap{
	"pkgPathRoot":  pkgPathRoot,
	"pkgPathSlug":  pkgPathSlug,
	"pkgPathBase":  pkgPathBase,
	"arrItemType":  stripArrBrackets,
	"mapKeyType":   mapKeyType,
	"mapValueType": mapValueType,
}

func pkgPathRoot(pkgPath string) string {
	if pkgPath == "" {
		return ""
	}

	parts := strings.Split(pkgPath, "/")

	return parts[0]
}

func pkgPathSlug(pkgPath string) string {
	if pkgPath == "" {
		return ""
	}

	pkgPath = strings.ReplaceAll(pkgPath, "-", "_")
	pkgPath = strings.ReplaceAll(pkgPath, "/", "__")

	return pkgPath
}

func pkgPathBase(pkgPath string) string {
	if pkgPath == "" {
		return ""
	}

	parts := strings.Split(pkgPath, "/")

	return parts[len(parts)-1]
}

func stripArrBrackets(typ string) string {
	return strings.TrimPrefix(typ, "[]")
}

func mapKeyType(typ string) string {
	typ = strings.TrimPrefix(typ, "map[")
	typRunes := []rune(typ)
	keyTypRunes := make([]rune, 0, len(typ))

	for i := 0; i < len(typ) && typRunes[i] != ']'; i++ {
		keyTypRunes = append(keyTypRunes, typRunes[i])
	}

	keyTyp := string(keyTypRunes)

	return keyTyp
}

func mapValueType(typ string) string {
	typ = strings.TrimPrefix(typ, "map[")
	typRunes := []rune(typ)
	var i int

	for i = 0; i < len(typRunes) && typRunes[i] != ']'; i++ {
		continue
	}

	valueTypRunes := make([]rune, 0, len(typ))

	for j := i + 1; j < len(typRunes); j++ {
		valueTypRunes = append(valueTypRunes, typRunes[j])
	}

	valueTyp := string(valueTypRunes)

	return valueTyp
}
