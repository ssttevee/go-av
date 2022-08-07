package main

import (
	"go/ast"
	"log"
	"strings"

	"github.com/dave/jennifer/jen"
)

func rawcname(ident *ast.Ident) (string, bool) {
	const prefix = "_Ctype_"

	if !strings.HasPrefix(ident.Name, prefix) {
		return ident.Name, false
	}

	return ident.Name[len(prefix):], true
}

func starsCodes(n int) []jen.Code {
	stars := make([]jen.Code, n)
	for i := 0; i < n; i++ {
		stars[i] = jen.Op("*")
	}

	return stars
}

func flattenPointers(expr ast.Expr) (stars []jen.Code, _ ast.Expr) {
	var n int
	for {
		starExpr, ok := expr.(*ast.StarExpr)
		if !ok {
			break
		}

		expr = starExpr.X
		n++
	}

	return starsCodes(n), expr
}

func goIntTypeToCIntType(name string) (string, bool) {
	switch name {
	case "uint32", "int32":
		return name + "_t", true

	case "uchar":
		return "uint8_t", true
	}

	return name, false
}

func ctypeFromGoExpr(expr ast.Expr) (string, bool) {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		if x, ok := e.X.(*ast.Ident); ok && x.Name == "unsafe" && e.Sel.Name == "Pointer" {
			return "void*", true
		}

	case *ast.ArrayType:
		if l, ok := e.Len.(*ast.BasicLit); ok && l.Value == "0" {
			if elt, ok := e.Elt.(*ast.Ident); ok && elt.Name == "byte" {
				return "void (*%s)()", true
			}
		}

	case *ast.StarExpr:
		cname, ok := ctypeFromGoExpr(e.X)
		if !ok {
			return "", false
		}

		if strings.HasSuffix(cname, ")") {
			return cname, true
		}

		return cname + "*", true

	case *ast.Ident:
		cname, ok := rawcname(e)
		if !ok {
			cname, ok = goIntTypeToCIntType(cname)
		}

		if !ok {
			// not a c type
			return cname, false
		}

		if cname == "uchar" {
			return "unsigned char", true
		}

		if cname == "uint" {
			return "unsigned", true
		}

		if strings.HasPrefix(cname, "struct_") {
			cname = "struct " + cname[7:]
		}

		return cname, true
	}

	log.Printf("unexpected expr type: %T", expr)
	return "", false
}

func isCVoid(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}

	cname, ok := rawcname(ident)
	if !ok {
		return false
	}

	return cname == "void"
}

func isCString(expr ast.Expr) bool {
	starExpr, ok := expr.(*ast.StarExpr)
	if !ok {
		return false
	}

	ident, ok := starExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	cname, ok := rawcname(ident)
	if !ok {
		return false
	}

	return cname == "char"
}
