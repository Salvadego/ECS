//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	// 1. Discover module path from go.mod
	modulePath := findModulePath()
	if modulePath == "" {
		fmt.Println("Could not find module path in go.mod")
		os.Exit(1)
	}

	// ECS import path is modulePath + /pkg/ecs
	ecsImport := modulePath + "/pkg/ecs"

	// Components directory
	const compDir = "internal/components"
	fset := token.NewFileSet()

	var names []string
	// 2. Scan folder for component types
	filepath.Walk(compDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		srcAST, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		for _, decl := range srcAST.Decls {
			if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
				for _, spec := range gd.Specs {
					ts := spec.(*ast.TypeSpec)
					// include struct types or aliases
					switch ts.Type.(type) {
					case *ast.StructType, *ast.Ident:
						names = append(names, ts.Name.Name)
						injectIDMethod(path, ts.Name.Name, ecsImport)
					}
				}
			}
		}
		return nil
	})

	// 3. Generate ecs_ids.go
	var buf bytes.Buffer
	buf.WriteString("package components\n\n")
	buf.WriteString("import \"" + ecsImport + "\"\n\n")
	buf.WriteString("const (\n")
	for _, name := range names {
		buf.WriteString(fmt.Sprintf("\t%sID ecs.ComponentID = iota\n", name))
	}
	buf.WriteString(")\n")
	os.WriteFile(filepath.Join(compDir, "components.go"), buf.Bytes(), 0644)
}

// findModulePath reads go.mod in cwd or parent to find module declaration
func findModulePath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	for dir := cwd; dir != "/"; dir = filepath.Dir(dir) {
		modFile := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(modFile); err == nil {
			lines := strings.SplitSeq(string(data), "\n")
			for l := range lines {
				if strings.HasPrefix(l, "module ") {
					return strings.TrimSpace(strings.TrimPrefix(l, "module "))
				}
			}
		}
	}
	return ""
}

// injectIDMethod opens the file (path), ensures ecs import, and appends ID() method
func injectIDMethod(path, comp, ecsImport string) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", path, err)
		return
	}
	if bytes.Contains(src, fmt.Appendf(nil, "func (c %s) ID", comp)) {
		return // already has ID() method
	}
	// Ensure ecs import
	if !bytes.Contains(src, []byte(ecsImport)) {
		// Insert import if necessary
		if idx := bytes.Index(src, []byte("import (")); idx >= 0 {
			parts := bytes.SplitN(src, []byte("import ("), 2)
			src = append(parts[0], []byte("import (\n\t\""+ecsImport+"\"\n")...)
			src = append(src, parts[1]...)
		} else {
			// No import block; add import after package
			lines := bytes.SplitN(src, []byte("\n"), 2)
			src = slices.Clone(lines[0])
			src = append(src, []byte("\nimport \""+ecsImport+"\"\n")...)
			src = append(src, lines[1]...)
		}
	}
	// Append ID method
	method := fmt.Sprintf("\nfunc (c %s) ID() ecs.ComponentID {\n\treturn %sID\n}\n", comp, comp)
	src = append(src, []byte(method)...)
	os.WriteFile(path, src, 0644)
}
