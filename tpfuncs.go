package main

import (
	"strings"
	"text/template"
)

var templateFuncs = template.FuncMap{
	"pkgPathRoot": pkgPathRoot,
	"pkgPathSlug": pkgPathSlug,
	"pkgPathBase": pkgPathBase,
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
