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

	stopOnlyFromMain := func(x *ast.FuncDecl) {
		if x.Name.Name == "main" || x.Body == nil {
			return
		}

		ast.Inspect(x.Body, func(node ast.Node) bool {

			expr, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := expr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			pkgIdent, ok := sel.X.(*ast.Ident)
			if !ok {
				return false
			}

			for _, f := range findArr {
				if pkgIdent.Name == f.pkg && sel.Sel.Name == f.name {
					pass.Reportf(expr.Pos(), "%s.%s found outside main function", f.pkg, f.name)
				}
			}

			return true
		})
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {

			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name == "main" && fn.Recv == nil || fn.Body == nil {
				continue
			}

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				switch x := node.(type) {
				case *ast.CallExpr: // выражение
					panicFunc(x)
				case *ast.FuncDecl: // выражение
					stopOnlyFromMain(x)
				}

				return true
			})
		}
	}

	return nil, nil
}
