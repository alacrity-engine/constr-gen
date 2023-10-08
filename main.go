package main

import (
	_ "embed"
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/golang-collections/collections/queue"
)

// TODO: sometimes a component requires
// a slice or a map of objects. I think
// they may be served in a prefab as
// []interface{} or map[interface{}]interface{}.
// I think I need to habdle such cases
// separately in field setters. I think
// I need to create a list() function for
// Lua scripts to pass arrays to the Go
// code as []interface{} instead of
// map[interface{}]interface{}.

var (
	gomodPath   string
	importBase  string
	gomodPrefix string

	//go:embed scene-component.go.tmpl
	sceneComponentTmpl string
	//go:embed scene-registry.go.tmpl
	sceneRegistryTmpl string
)

// parseFlags parses the command line
// arguments passed as flag values.
func parseFlags() {
	flag.StringVar(&gomodPath, "gomod", "/home/zergon321/go/src/test-game",
		"Go module the game module is related to")

	flag.Parse()
}

// getImportBase obtains the Go module
// part of the game module package path for
// writing imports in the autogenerated file.
func getImportBase() {
	origWD, err := os.Getwd()
	handleError(err)
	err = os.Chdir(gomodPath)
	handleError(err)
	gomodDir, err := os.Getwd()
	handleError(err)
	err = os.Chdir(origWD)
	handleError(err)

	// Make paths absolute.
	gomodPath, err = filepath.Abs(gomodPath)
	handleError(err)

	gomodPath := path.Dir(gomodDir)
	importBase = strings.TrimPrefix(gomodDir, gomodPath+"/")
	gomodPrefix = gomodPath
}

func main() {
	parseFlags()
	getImportBase()

	// Get all the files and directories of the module.
	entries, err := os.ReadDir(gomodPath)
	handleError(err)
	traverseQueue := queue.New()

	// Enqueue them for breadth-first traversal.
	for _, entry := range entries {
		tracker := FileTracker{
			dir:  gomodPath,
			info: entry,
		}

		traverseQueue.Enqueue(tracker)
	}

	// Collect all the package names
	// to write imports.
	packages := map[string]struct{}{}
	// Store all the type declarations
	// that are Alacrity components.
	comps := []TypeDeclaration{}

	for traverseQueue.Len() > 0 {
		entry := traverseQueue.Dequeue().(FileTracker)

		// If the entry is a directory.
		if entry.info.IsDir() {
			// Read the directory for entries.
			dirName := path.Join(entry.dir, entry.info.Name())
			entries, err = os.ReadDir(dirName)
			handleError(err)

			// Enqueue them for breadth-first traversal.
			for _, entry := range entries {
				tracker := FileTracker{
					dir:  dirName,
					info: entry,
				}

				traverseQueue.Enqueue(tracker)
			}

			continue
		}

		if !strings.HasSuffix(entry.info.Name(), ".go") {
			continue
		}

		// Analyze the source code file
		// using Go AST tools to find all
		// the type declarations.
		fileName := path.Join(entry.dir, entry.info.Name())
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, fileName,
			nil, parser.ParseComments)
		handleError(err)

		imports := map[string]struct{}{}

		for _, imp := range file.Imports {
			importPath := imp.Path.Value

			if imp.Name != nil {
				importPath = imp.Name.Name + " " + importPath
			}

			imports[importPath] = struct{}{}
		}

		typeDecls := []TypeDeclaration{}
		packageName := file.Name.Name
		packages[packageName] = struct{}{}
		compPkg := strings.TrimPrefix(entry.dir, gomodPrefix+"/")

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
							typeDecl := TypeDeclaration{
								packageName: packageName,
								typeName:    typeName,
								fieldList:   fieldList,
								packagePath: compPkg,
							}

							typeDecls = append(typeDecls, typeDecl)
						}
					}
				}
			}

			return true
		})

		// Select only the types that are components.
		for _, typeDecl := range typeDecls {
			// Find the "engine.BaseComponent".
			for _, field := range typeDecl.fieldList {
				t, ok := field.Type.(*ast.SelectorExpr)

				if !ok {
					continue
				}

				x, ok := t.X.(*ast.Ident)

				if !ok {
					continue
				}

				if t.Sel.Name == "BaseComponent" && x.Name == "engine" {
					typeDecl.imports = imports
					comps = append(comps, typeDecl)

					break
				}
			}
		}
	}

	// Generate template data.
	templateComps := make([]ComponentTemplateData, 0, len(comps))
	compPkgs := map[string]struct{}{}

	for _, comp := range comps {
		templateComp := ComponentTemplateData{
			Imports: comp.imports,
			Name:    comp.typeName,
			PkgPath: comp.packagePath,
			Fields:  make([]ComponentFieldTemplateData, 0, len(comp.fieldList)),
		}

		compPkgs["\""+templateComp.PkgPath+"\""] = struct{}{}

		for _, field := range comp.fieldList {
			// If the field is unnamed, skip it.
			if len(field.Names) <= 0 {
				continue
			}

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
						continue
					}

					paramTypeName = "[]" + x.Name +
						"." + tt.Sel.Name

				default:
					continue
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
						continue
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
						continue
					}

					valueType = x.Name + "." + tt.Sel.Name
				}

				paramTypeName = "map[" + keyType + "]" + valueType

			// If it's a structure from some package
			// (such as time.Duration or pixel.Vec).
			case *ast.SelectorExpr:
				x, ok := t.X.(*ast.Ident)

				if !ok {
					continue
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
						continue
					}

					paramTypeName = "*" + x.Name + "." + inner.Sel.Name
				}

			default:
				continue
			}

			paramName := field.Names[0].Name

			// If the field is not public, skip it.
			if field.Tag == nil || !strings.Contains(field.Tag.Value, `iris:"exported"`) {
				continue
			}

			field := ComponentFieldTemplateData{
				Name: paramName,
				Type: paramTypeName,
			}
			templateComp.Fields = append(templateComp.Fields, field)
		}

		templateComps = append(templateComps, templateComp)
		_ = templateComps
	}

	// Write the source files to the
	// main package of the module.
	compTemplate := template.New("scene-component.go.tmpl")
	_, err = compTemplate.Funcs(templateFuncs).
		Parse(sceneComponentTmpl)
	handleError(err)

	for _, templateComp := range templateComps {
		fpath := path.Join(gomodPrefix, templateComp.PkgPath, "autogen."+templateComp.Name+".go")
		file, err := os.Create(fpath)
		handleError(err)

		err = compTemplate.Execute(file, map[string]interface{}{
			"moduleRootPath":    importBase,
			"pkgPath":           templateComp.PkgPath,
			"componentTypeName": templateComp.Name,
			"fields":            templateComp.Fields,
			"imports":           templateComp.Imports,
		})
		handleError(err)

		err = file.Close()
		handleError(err)

		// Add missing imports using 'goimports'.
		editedSource, err := exec.Command("goimports", fpath).Output()
		handleError(err)
		// Write the edited source to the autogenerated file.
		file, err = os.OpenFile(fpath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		handleError(err)
		_, err = file.WriteString(string(editedSource))
		handleError(err)

		err = file.Close()
		handleError(err)
	}

	regTemplate := template.New("scene-registry.go.tmpl")
	_, err = regTemplate.Funcs(templateFuncs).
		Parse(sceneRegistryTmpl)
	handleError(err)

	fpath := path.Join(gomodPath, "registry.go")
	file, err := os.Create(fpath)
	handleError(err)

	err = regTemplate.Execute(file, map[string]interface{}{
		"moduleRootPath": importBase,
		"compTypes":      templateComps,
		"pkgImports":     compPkgs,
	})
	handleError(err)

	err = file.Close()
	handleError(err)

	// Add missing imports using 'goimports'.
	editedSource, err := exec.Command("goimports", fpath).Output()
	handleError(err)
	// Write the edited source to the autogenerated file.
	file, err = os.OpenFile(fpath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	handleError(err)
	_, err = file.WriteString(string(editedSource))
	handleError(err)

	err = file.Close()
	handleError(err)
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
