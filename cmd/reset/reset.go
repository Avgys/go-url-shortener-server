package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"text/template"
)

type templateDesc struct {
	Package string
	Types   []structDesc
}

type structDesc struct {
	Type        string
	FieldResets []string
}

type fieldDesc struct {
	Receiver  string
	Name      string
	Value     string
	TypeName  string
	InnerCode string
}

func main() {
	fname := os.Getenv("GOFILE")
	if fname == "" {
		fail("GOFILE environment variable is not set")
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fname, nil, parser.ParseComments)
	if err != nil {
		fail("parse %s: %v", fname, err)
	}

	resetable := collectResetableTypes(f)
	structDescs := buildTree(fset, f, resetable)

	code := prepareCode(structDescs, f.Name.Name)

	if err := storeToFile(code); err != nil {
		fail("write reset.gen.go: %v", err)
	}
}

func buildTree(fset *token.FileSet, f *ast.File, resetable map[string]struct{}) []structDesc {
	entries := make([]structDesc, 0)

	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}

		for _, s := range gd.Specs {
			ts, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if !isNeedToReset(gd.Doc) && !isNeedToReset(ts.Doc) {
				continue
			}

			structType, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structD := structDesc{Type: ts.Name.Name}
			structD.FieldResets = getFieldsResetCode(fset, structType, resetable)
			entries = append(entries, structD)
		}
	}

	return entries
}

func getFieldsResetCode(fset *token.FileSet, structType *ast.StructType, resetable map[string]struct{}) []string {
	fields := make([]string, 0)

	for _, field := range structType.Fields.List {
		names := field.Names
		if len(names) == 0 {
			name := embeddedName(field.Type)
			if name == "" {
				warn("skip embedded field: unsupported type %s", typeExprString(fset, field.Type))
				continue
			}
			names = []*ast.Ident{{Name: name}}
		}

		for _, n := range names {
			resetCode := getResetCode(fset, n.Name, field.Type, resetable)
			if resetCode == "" {
				warn("skip field %s: unsupported type %s", n.Name, typeExprString(fset, field.Type))
				continue
			}
			fields = append(fields, resetCode)
		}
	}

	return fields
}

func prepareCode(types []structDesc, pkg string) []byte {
	var buf bytes.Buffer

	err := mainTmpl.Execute(&buf, templateDesc{Package: pkg, Types: types})
	if err != nil {
		fail("execute main template: %v", err)
	}

	code, err := format.Source(buf.Bytes())
	if err != nil {
		fail("format generated code: %v", err)
	}

	return code
}

func collectResetableTypes(f *ast.File) map[string]struct{} {
	out := make(map[string]struct{})

	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}

		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if !isNeedToReset(gd.Doc) && !isNeedToReset(ts.Doc) {
				continue
			}

			if _, ok := ts.Type.(*ast.StructType); !ok {
				continue
			}

			out[ts.Name.Name] = struct{}{}
		}
	}

	return out
}

func isNeedToReset(cg *ast.CommentGroup) bool {
	const subStr = "// generate:reset"

	if cg == nil {
		return false
	}

	for _, c := range cg.List {
		if strings.Contains(c.Text, subStr) {
			return true
		}
	}

	return false
}

func storeToFile(code []byte) error {
	return os.WriteFile("reset.gen.go", code, 0644)
}

func getResetCode(fset *token.FileSet, name string, expr ast.Expr, resetable map[string]struct{}) string {
	decl := fieldDesc{Receiver: "v", Name: name}

	switch t := expr.(type) {
	case *ast.StarExpr:
		return pointerResetCode(fset, expr, decl, resetable)
	case *ast.ChanType:
		return applyResetTemplate(chanCloseTmpl, decl)
	case *ast.MapType:
		return applyResetTemplate(mapTmpl, decl)
	case *ast.ArrayType:
		if t.Len == nil {
			return applyResetTemplate(arrayTmpl, decl)
		}
		decl.TypeName = typeExprString(fset, expr)
		return applyResetTemplate(arrayZeroTmpl, decl)
	case *ast.Ident:
		return identResetCode(t, decl, resetable)
	case *ast.SelectorExpr:
		decl.TypeName = typeExprString(fset, t)
		if isResetableType(t.Sel.Name, resetable) {
			return applyResetTemplate(callResetTmpl, decl)
		}
		return applyResetTemplate(zeroStructTmpl, decl)
	case *ast.StructType:
		decl.TypeName = typeExprString(fset, t)
		return applyResetTemplate(zeroStructTmpl, decl)
	case *ast.InterfaceType:
		decl.Value = "nil"
		return applyResetTemplate(valueTmpl, decl)
	default:
		return ""
	}
}

func identResetCode(ident *ast.Ident, decl fieldDesc, resetable map[string]struct{}) string {
	if zero, ok := defaultValueByType[ident.Name]; ok {
		decl.Value = zero
		return applyResetTemplate(valueTmpl, decl)
	}

	if isResetableType(ident.Name, resetable) {
		return applyResetTemplate(callResetTmpl, decl)
	}

	decl.TypeName = ident.Name
	return applyResetTemplate(zeroStructTmpl, decl)
}

func isResetableType(name string, resetable map[string]struct{}) bool {
	_, ok := resetable[name]
	return ok
}

func pointerResetCode(fset *token.FileSet, expr ast.Expr, decl fieldDesc, resetable map[string]struct{}) string {
	depth := 0
	inner := expr

	for {
		star, ok := inner.(*ast.StarExpr)
		if !ok {
			break
		}
		depth++
		inner = star.X
	}

	if depth == 0 {
		return ""
	}

	field := decl.Receiver + "." + decl.Name
	innerStmt := buildInnerPointerStmt(fset, inner, field, depth, resetable)
	if innerStmt == "" {
		return ""
	}

	decl.InnerCode = buildPointerNilChecks(decl.Receiver, decl.Name, depth, innerStmt)
	return applyResetTemplate(pointerResetTmpl, decl)
}

func buildInnerPointerStmt(fset *token.FileSet, inner ast.Expr, field string, depth int, resetable map[string]struct{}) string {
	deref := strings.Repeat("*", depth) + field

	switch t := inner.(type) {
	case *ast.Ident:
		if zero, ok := defaultValueByType[t.Name]; ok {
			return fmt.Sprintf("%s = %s", deref, zero)
		}

		if isResetableType(t.Name, resetable) {
			return fmt.Sprintf("%s.Reset()", field)
		}

		return fmt.Sprintf("%s = %s{}", deref, t.Name)

	case *ast.SelectorExpr:
		typeName := typeExprString(fset, t)
		if isResetableType(t.Sel.Name, resetable) {
			return fmt.Sprintf("%s.Reset()", field)
		}
		return fmt.Sprintf("%s = %s{}", deref, typeName)

	case *ast.StructType:
		typeName := typeExprString(fset, t)
		return fmt.Sprintf("%s = %s{}", deref, typeName)

	case *ast.MapType:
		return guardNilTarget(deref, fmt.Sprintf("clear(%s)", deref))

	case *ast.ArrayType:
		if t.Len == nil {
			return fmt.Sprintf("%s = %s[:0]", deref, deref)
		}
		typeName := typeExprString(fset, inner)
		return fmt.Sprintf("%s = %s{}", deref, typeName)

	case *ast.ChanType:
		return guardNilTarget(deref, fmt.Sprintf("close(%s)", deref)+"\n\t\t"+fmt.Sprintf("%s = nil", deref))

	case *ast.InterfaceType:
		return fmt.Sprintf("%s = nil", deref)

	default:
		return ""
	}
}

func guardNilTarget(target, stmt string) string {
	return fmt.Sprintf("if %s != nil {\n\t\t%s\n\t}", target, stmt)
}

func buildPointerNilChecks(recv, name string, depth int, innerStmt string) string {
	field := recv + "." + name

	if depth <= 1 {
		return "\t" + innerStmt
	}

	lines := make([]string, 0, depth*2)

	for i := 1; i < depth; i++ {
		stars := strings.Repeat("*", i)
		lines = append(lines, strings.Repeat("\t", i)+fmt.Sprintf("if %s%s != nil {", stars, field))
	}

	lines = append(lines, strings.Repeat("\t", depth)+innerStmt)

	for i := depth - 1; i >= 1; i-- {
		lines = append(lines, strings.Repeat("\t", i)+"}")
	}

	return strings.Join(lines, "\n")
}

func applyResetTemplate(tmpl *template.Template, decl fieldDesc) string {
	var buf bytes.Buffer

	err := tmpl.Execute(&buf, decl)
	if err != nil {
		fail("execute template: %v", err)
	}

	return buf.String()
}

func typeExprString(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, expr); err != nil {
		return ""
	}
	return buf.String()
}

var defaultValueByType = map[string]string{
	"string":     `""`,
	"bool":       "false",
	"error":      "nil",
	"int":        "0",
	"int8":       "0",
	"int16":      "0",
	"int32":      "0",
	"int64":      "0",
	"uint":       "0",
	"uint8":      "0",
	"uint16":     "0",
	"uint32":     "0",
	"uint64":     "0",
	"uintptr":    "0",
	"byte":       "0",
	"rune":       "0",
	"float32":    "0",
	"float64":    "0",
	"complex64":  "0",
	"complex128": "0",
}

func embeddedName(typeExpr ast.Expr) string {
	for {
		star, ok := typeExpr.(*ast.StarExpr)
		if !ok {
			break
		}
		typeExpr = star.X
	}

	switch t := typeExpr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return t.Sel.Name
	default:
		return ""
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "reset: "+format+"\n", args...)
	os.Exit(1)
}

func warn(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "reset: warning: "+format+"\n", args...)
}
