package main

import (
	"go/ast"
	"strings"

	"github.com/dave/jennifer/jen"
)

func rawcname(ident *ast.Ident) (string, bool) {
	const prefix = "_Ctype_"

	if !strings.HasPrefix(ident.Name, prefix) {
		return "", false
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
