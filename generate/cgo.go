package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
)

func rendertemp(f *jen.File) (string, error) {
	var buf bytes.Buffer
	if err := f.Render(&buf); err != nil {
		return "", errors.WithStack(err)
	}

	w, err := ioutil.TempFile("", "*.go")
	if err != nil {
		return "", errors.WithStack(err)
	}

	defer w.Close()

	if _, err := buf.WriteTo(w); err != nil {
		_ = os.Remove(w.Name())
		return "", errors.WithStack(err)
	}

	return w.Name(), nil
}

func gotypes(filename string) (*ast.File, error) {
	dirname, err := ioutil.TempDir("", "*")
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer os.RemoveAll(dirname)

	cmd := exec.Command("go", "tool", "cgo", "-objdir", dirname, filename)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, errors.WithStack(err)
	}

	source, err := ioutil.ReadFile(path.Join(dirname, "_cgo_gotypes.go"))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	t, err := parser.ParseFile(token.NewFileSet(), "", source, 0)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return t, nil
}

func (u *Unit) CgoAST() (*ast.File, error) {
	if u.cachedCgoAST != nil {
		return u.cachedCgoAST, nil
	}

	f := jen.NewFile("main")
	f.CgoPreamble(u.CgoPreamble)

	for _, ct := range u.ConvertTypes {
		f.Type().Id("T").Qual("C", ct.CName)
	}

	for _, ct := range u.WrapFuncs {
		f.Var().Id("f").Op("=").Qual("C", ct.CName).Params()
	}

	tempfilename, err := rendertemp(f)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer os.Remove(tempfilename)

	t, err := gotypes(tempfilename)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	u.cachedCgoAST = t

	return t, nil
}

type TypeSpec struct {
	Type string
}

type StructSpec struct {
	Fields []*ast.Field
}

func (u *Unit) ctypeSpec(cname string) (interface{}, error) {
	t, err := u.CgoAST()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	obj := t.Scope.Lookup("_Ctype_" + cname)
	if obj == nil {
		return nil, errors.Errorf("type not defined: %s", cname)
	}

	typeSpec, ok := obj.Decl.(*ast.TypeSpec)
	if !ok {
		return nil, errors.Errorf("unexpected decl type: %T (%s)", obj.Decl, cname)
	}

	switch typ := typeSpec.Type.(type) {
	case *ast.StructType:
		return &StructSpec{
			Fields: typ.Fields.List,
		}, nil

	case *ast.Ident:
		return &TypeSpec{
			Type: typ.Name,
		}, nil

	default:
		if sel, ok := typeSpec.Type.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				if ident.Name == "_cgopackage" {
					if sel.Sel.Name == "Incomplete" {
						return &TypeSpec{
							Type: sel.Sel.Name,
						}, nil
					}
				}
			}
		}

		return nil, errors.Errorf("unexpected type type: %T (%s)", typ, cname)
	}
}

type FuncSpec struct {
	Params     []*ast.Field
	ResultType ast.Expr
}

func (u *Unit) cfuncSpec(cname string) (*FuncSpec, error) {
	t, err := u.CgoAST()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	obj := t.Scope.Lookup("_Cfunc_" + cname)
	if obj == nil {
		return nil, errors.New("func not defined")
	}

	funcDecl, ok := obj.Decl.(*ast.FuncDecl)
	if !ok {
		return nil, errors.Errorf("unexpected decl type: %T", obj.Decl)
	}

	var resultType ast.Expr
	for _, f := range funcDecl.Type.Results.List {
		for range f.Names {
			if resultType != nil {
				return nil, errors.Errorf("too many return types: %s", cname)
			}

			resultType = f.Type
		}
	}

	return &FuncSpec{
		Params:     funcDecl.Type.Params.List,
		ResultType: resultType,
	}, nil
}
