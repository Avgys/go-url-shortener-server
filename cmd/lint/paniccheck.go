package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var PanicCheckAnalyzer = &analysis.Analyzer{
	Name: "paniccheck",
	Doc:  "check if panic called outside main func",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	panicFunc := func(x *ast.CallExpr) {
		if id, ok := x.Fun.(*ast.Ident); ok {
			if id.Name == "panic" {
				pass.Reportf(id.Pos(), "panic called here")
			}
		}
	}

	findArr := []struct {
		pkg  string
		name string
	}{
		{pkg: "log", name: "Fatal"},
		{pkg: "os", name: "Exit"},
	}

	stopOnlyFromMain := func(x *ast.CallExpr) {
		sel, ok := x.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		pkgIdent, ok := sel.X.(*ast.Ident)
		if !ok {
			return
		}

		for _, f := range findArr {
			if pkgIdent.Name == f.pkg && sel.Sel.Name == f.name {
				pass.Reportf(x.Pos(), "%s.%s found outside main function", f.pkg, f.name)
			}
		}
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {

			fn, ok := decl.(*ast.FuncDecl)
			if !ok || pass.Pkg.Name() == "main" && fn.Recv == nil || fn.Body == nil {
				continue
			}

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				switch x := node.(type) {
				case *ast.CallExpr: // выражение
					panicFunc(x)
					stopOnlyFromMain(x)
				}

				return true
			})
		}
	}

	return nil, nil
}
