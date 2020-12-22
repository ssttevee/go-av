package main

import (
	"bytes"
	"go/ast"
	"go/token"
	"io/ioutil"
	"path"
	"strconv"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
	"github.com/ssttevee/go-caseconv"
)

func (u *Unit) fieldType(cname string, fieldName string) *GoRef {
	fields, ok := u.FieldTypes[cname]
	if !ok {
		return nil
	}

	ref, ok := fields[fieldName]
	if !ok {
		return nil
	}

	return &ref
}

func (u *Unit) paramType(cname string, index int) *GoRef {
	params, ok := u.ParamTypes[cname]
	if !ok {
		return nil
	}

	ref, ok := params[index]
	if !ok {
		return nil
	}

	return &ref
}

type generateContext struct {
	name  string
	ct2gt map[string]*ConvertType
}

func (ctx *generateContext) castFromC(ctype ast.Expr) (cast func(jen.Code) jen.Code, needsAssign bool) {
	stars, ctype := flattenPointers(ctype)
	ident, ok := ctype.(*ast.Ident)
	if !ok {
		return nil, false
	}

	rawname, ok := rawcname(ident)
	if !ok {
		return nil, false
	}

	gotype, ok := ctx.ct2gt[rawname]
	if !ok {
		return nil, false
	}

	if len(stars) > 0 {
		return func(expr jen.Code) jen.Code {
			return jen.Parens(jen.Add(stars...).Add(gotype.Code())).Parens(jen.Qual("unsafe", "Pointer").Parens(expr))
		}, false
	}

	return func(expr jen.Code) jen.Code {
		return jen.Op("*").Parens(jen.Op("*").Add(gotype.Code())).Parens(jen.Qual("unsafe", "Pointer").Parens(jen.Op("&").Add(expr)))
	}, true
}

func (ctx *generateContext) castToC(ctype ast.Expr) (cast func(jen.Code) jen.Code, needsAssign bool) {
	stars, ctype := flattenPointers(ctype)
	ident, ok := ctype.(*ast.Ident)
	if !ok {
		return nil, false
	}

	rawname, ok := rawcname(ident)
	if !ok {
		return nil, false
	}

	if _, ok := ctx.ct2gt[rawname]; !ok {
		return nil, false
	}

	if len(stars) > 0 {
		return func(expr jen.Code) jen.Code {
			return jen.Parens(jen.Add(stars...).Qual("C", rawname)).Parens(jen.Qual("unsafe", "Pointer").Parens(expr))
		}, false
	}

	return func(expr jen.Code) jen.Code {
		return jen.Op("*").Parens(jen.Op("*").Qual("C", rawname)).Parens(jen.Qual("unsafe", "Pointer").Parens(jen.Op("&").Add(expr)))
	}, true
}

func (ctx *generateContext) renderExpr(expr ast.Expr, isType bool) (jen.Code, error) {
	switch expr := expr.(type) {
	case *ast.Ident:
		if rawname, ok := rawcname(expr); ok {
			gotype, ok := ctx.ct2gt[rawname]
			if ok {
				return gotype.Code(), nil
			}

			return jen.Qual("C", rawname), nil
		}

		return jen.Id(expr.Name), nil

	case *ast.ArrayType:
		el, err := ctx.renderExpr(expr.Elt, isType)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		var indices []jen.Code
		if expr.Len != nil {
			l, err := ctx.renderExpr(expr.Len, isType)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			indices = append(indices, l)
		}

		return jen.Index(indices...).Add(el), nil

	case *ast.StarExpr:
		x, err := ctx.renderExpr(expr.X, isType)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return jen.Op("*").Add(x), nil

	case *ast.BasicLit:
		switch expr.Kind {
		case token.INT:
			n, err := strconv.ParseInt(expr.Value, 10, 64)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			return jen.Lit(int(n)), nil

		case token.FLOAT:
			n, err := strconv.ParseFloat(expr.Value, 64)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			return jen.Lit(n), nil

		case token.IMAG:
			n, err := strconv.ParseComplex(expr.Value, 64)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			return jen.Lit(n), nil

		case token.CHAR:
			return jen.LitRune([]rune(expr.Value)[0]), nil

		case token.STRING:
			return jen.Lit(expr.Value), nil

		default:
			return nil, errors.Errorf("unexpected literal kind: %s", expr.Kind)
		}

	case *ast.SelectorExpr:
		if isType {
			ident, ok := expr.X.(*ast.Ident)
			if !ok {
				return nil, errors.Errorf("expected *ast.Ident but got %T", expr.X)
			}

			return jen.Qual(ident.Name, expr.Sel.Name), nil
		}

		x, err := ctx.renderExpr(expr.X, false)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return jen.Add(x).Dot(expr.Sel.Name), nil

	default:
		return nil, errors.Errorf("unexpected expr type: %T", expr)
	}
}

func (ctx *generateContext) generateFuncWrapper(u *Unit, wf *WrapFunc) (jen.Code, error) {
	spec, err := u.cfuncSpec(wf.CName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var returnTypes []jen.Code
	if spec.ResultType != nil {
		if !isCVoid(spec.ResultType) {
			typeCode, err := ctx.renderExpr(spec.ResultType, true)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			returnTypes = append(returnTypes, jen.Parens(typeCode))
		}
	}

	var paramCodes []jen.Code
	for i, param := range spec.Params {
		var typeCode jen.Code
		if override := u.paramType(wf.CName, i); override != nil {
			typeCode = override.Code()
		} else if isCString(param.Type) {
			typeCode = jen.String()
		} else if typeCode, err = ctx.renderExpr(param.Type, true); err != nil {
			return nil, errors.WithStack(err)
		}

		for _, name := range param.Names {
			paramCodes = append(paramCodes, jen.Id(caseconv.CamelCase(name.Name)).Add(typeCode))
		}
	}

	var funcStmts []jen.Code
	callParams := make([]jen.Code, 0, len(spec.Params))
	var i int
	for _, param := range spec.Params {
		for _, name := range param.Names {
			var paramCode jen.Code = jen.Id(name.Name)
			if override := u.paramType(wf.CName, i); override != nil {
				exprCode, err := ctx.renderExpr(param.Type, true)
				if err != nil {
					return nil, errors.WithStack(err)
				}

				paramCode = jen.Parens(exprCode).Parens(paramCode)
			} else if isCString(param.Type) {
				csident := jen.Id("s" + strconv.FormatInt(int64(len(callParams)), 10))
				funcStmts = append(funcStmts, jen.Var().Add(csident).Op("*").Qual("C", "char"))
				funcStmts = append(funcStmts, jen.If(jen.Add(paramCode).Op("!=").Lit("")).Block(
					jen.Add(csident).Op("=").Qual("C", "CString").Params(paramCode),
					jen.Defer().Qual("C", "free").Params(jen.Qual("unsafe", "Pointer").Params(csident)),
				))
				paramCode = csident
			} else if cast, _ := ctx.castToC(param.Type); cast != nil {
				funcStmts = append(funcStmts, jen.Defer().Qual("runtime", "KeepAlive").Parens(paramCode))
				paramCode = cast(paramCode)
			}

			callParams = append(callParams, paramCode)
			i++
		}
	}

	call := jen.Qual("C", wf.CName).Params(callParams...)

	if len(returnTypes) == 0 {
		funcStmts = append(funcStmts, call)
	} else {
		var ret jen.Code
		if cast, needsAssign := ctx.castFromC(spec.ResultType); cast == nil {
			ret = call
		} else if needsAssign {
			funcStmts = append(funcStmts, jen.Id("ret").Op(":=").Add(call))
			ret = cast(jen.Id("ret"))
		} else {
			ret = cast(call)
		}

		funcStmts = append(funcStmts, jen.Return(ret))
	}

	return jen.Func().Id(wf.GoName).Params(paramCodes...).Add(returnTypes...).Block(funcStmts...), nil
}

func generate(ctx *generateContext, filename string, u *Unit) error {
	f := jen.NewFile(ctx.name)
	f.CgoPreamble(u.CgoPreamble)
	f.HeaderComment("Code generated by robots; DO NOT EDIT.")

	var shouldOutput bool
	for _, convertType := range u.ConvertTypes {
		if convertType.GoPackage != "" {
			continue
		}

		spec, err := u.ctypeSpec(convertType.CName)
		if err != nil {
			return errors.WithStack(err)
		}

		switch spec := spec.(type) {
		case *StructSpec:
			var fields []jen.Code
			for _, field := range spec.Fields {
				for _, name := range field.Names {
					var typeCode jen.Code
					if override := u.fieldType(convertType.CName, name.Name); override != nil {
						typeCode = override.Code()
					} else if typeCode, err = ctx.renderExpr(field.Type, true); err != nil {
						return errors.WithStack(err)
					}

					identName := caseconv.PascalCase(name.Name)
					if identName == "" {
						identName = name.Name
					}

					fields = append(fields, jen.Id(identName).Add(typeCode))
				}
			}

			f.Type().Id(convertType.GoName).Struct(fields...)

			shouldOutput = true

		case *TypeSpec:
		default:
			return errors.Errorf("unexpected spec type: %T", spec)
		}
	}

	for _, wf := range u.WrapFuncs {
		code, err := ctx.generateFuncWrapper(u, wf)
		if err != nil {
			return errors.WithStack(err)
		}

		if code != nil {
			f.Add(code)
		}

		shouldOutput = true
	}

	if !shouldOutput {
		return nil
	}

	var buf bytes.Buffer
	if err := f.Render(&buf); err != nil {
		return errors.WithStack(err)
	}

	ext := path.Ext(filename)
	if err := ioutil.WriteFile(filename[:len(filename)-len(ext)]+"_gen"+ext, buf.Bytes(), 0666); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
