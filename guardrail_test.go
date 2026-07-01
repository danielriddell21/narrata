package narrata

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// forbidden lists concepts that must never appear in the public narration API.
// See architecture.md §11 and spec.md §5.
var forbidden = []string{
	"Assistant", "ChatSession", "MemoryStore", "Tool", "ToolRegistry",
	"Agent", "Planner", "Retriever", "VectorStore", "Browser", "Scheduler",
	"Chat", "Remember", "Plan", "CallTool",
}

// TestNoForbiddenExportedIdentifiers ensures the public package declares none of
// the agent/assistant concepts Narrata explicitly excludes.
func TestNoForbiddenExportedIdentifiers(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, 0) //nolint:staticcheck // single-package guardrail scan; build-tag handling is irrelevant here
	if err != nil {
		t.Fatalf("parsing package: %v", err)
	}

	banned := make(map[string]bool, len(forbidden))
	for _, f := range forbidden {
		banned[f] = true
	}

	for _, pkg := range pkgs {
		if strings.HasSuffix(pkg.Name, "_test") {
			continue
		}
		for fileName, file := range pkg.Files {
			if strings.HasSuffix(fileName, "_test.go") {
				continue
			}
			for _, decl := range file.Decls {
				for _, name := range exportedNames(decl) {
					if banned[name] {
						t.Errorf("forbidden exported identifier %q declared in %s", name, fileName)
					}
				}
			}
		}
	}
}

func exportedNames(decl ast.Decl) []string {
	var names []string
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if d.Recv == nil && d.Name.IsExported() {
			names = append(names, d.Name.Name)
		}
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				if s.Name.IsExported() {
					names = append(names, s.Name.Name)
				}
			case *ast.ValueSpec:
				for _, n := range s.Names {
					if n.IsExported() {
						names = append(names, n.Name)
					}
				}
			}
		}
	}
	return names
}
