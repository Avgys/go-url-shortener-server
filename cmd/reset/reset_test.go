package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func parseStructField(t *testing.T, src, fieldName string) (ast.Expr, map[string]struct{}) {
	t.Helper()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "fixture.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	resetable := collectResetableTypes(f)

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

			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}

			for _, field := range st.Fields.List {
				for _, n := range field.Names {
					if n.Name == fieldName {
						return field.Type, resetable
					}
				}
			}
		}
	}

	t.Fatalf("field %s not found", fieldName)
	return nil, nil
}

func TestGetResetCode_slice(t *testing.T) {
	fset := token.NewFileSet()
	expr, resetable := parseStructField(t, `package p
// generate:reset
type S struct { items []int }`, "items")

	got := getResetCode(fset, "items", expr, resetable)
	if !strings.Contains(got, "v.items = v.items[:0]") {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestGetResetCode_array(t *testing.T) {
	fset := token.NewFileSet()
	expr, resetable := parseStructField(t, `package p
// generate:reset
type S struct { buf [4]byte }`, "buf")

	got := getResetCode(fset, "buf", expr, resetable)
	if !strings.Contains(got, "v.buf = [4]byte{}") {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestGetResetCode_pointerString(t *testing.T) {
	fset := token.NewFileSet()
	expr, resetable := parseStructField(t, `package p
// generate:reset
type S struct { strP *string }`, "strP")

	got := getResetCode(fset, "strP", expr, resetable)
	if !strings.Contains(got, "if v.strP != nil") || !strings.Contains(got, "*v.strP = \"\"") {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestGetResetCode_pointerMapNilGuard(t *testing.T) {
	fset := token.NewFileSet()
	expr, resetable := parseStructField(t, `package p
// generate:reset
type S struct { m *map[string]string }`, "m")

	got := getResetCode(fset, "m", expr, resetable)
	if !strings.Contains(got, "if *v.m != nil") || !strings.Contains(got, "clear(*v.m)") {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestGetResetCode_channelIdempotent(t *testing.T) {
	fset := token.NewFileSet()
	expr, resetable := parseStructField(t, `package p
// generate:reset
type S struct { ch chan int }`, "ch")

	got := getResetCode(fset, "ch", expr, resetable)
	if !strings.Contains(got, "close(v.ch)") || !strings.Contains(got, "v.ch = nil") {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestGetResetCode_selectorType(t *testing.T) {
	fset := token.NewFileSet()
	expr, resetable := parseStructField(t, `package p
import "time"
// generate:reset
type S struct { when time.Time }`, "when")

	got := getResetCode(fset, "when", expr, resetable)
	if !strings.Contains(got, "v.when = time.Time{}") {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestCollectResetableTypes_typeSpecDoc(t *testing.T) {
	src := `package p
type (
	// generate:reset
	Foo struct { x int }
)`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "fixture.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	resetable := collectResetableTypes(f)
	if _, ok := resetable["Foo"]; !ok {
		t.Fatalf("expected Foo in resetable set, got %v", resetable)
	}
}
