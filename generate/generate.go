package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"log"
	"path"
	"sort"
	"strconv"
	"strings"

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

func (ctx *generateContext) renderExpr(expr ast.Expr, isType bool, altcname func(string) (string, bool), preferC bool) (jen.Code, error) {
	switch expr := expr.(type) {
	case *ast.Ident:
		rawname, ok := rawcname(expr)
		if !ok && altcname != nil {
			rawname, ok = altcname(rawname)
		}

		if ok {
			gotype, ok := ctx.ct2gt[rawname]
			if ok && !preferC {
				return gotype.Code(), nil
			}

			return jen.Qual("C", rawname), nil
		}

		return jen.Id(expr.Name), nil

	case *ast.ArrayType:
		el, err := ctx.renderExpr(expr.Elt, isType, altcname, preferC)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		var indices []jen.Code
		if expr.Len != nil {
			l, err := ctx.renderExpr(expr.Len, isType, altcname, preferC)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			indices = append(indices, l)
		}

		return jen.Index(indices...).Add(el), nil

	case *ast.StarExpr:
		x, err := ctx.renderExpr(expr.X, isType, altcname, preferC)
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

		x, err := ctx.renderExpr(expr.X, false, altcname, preferC)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		return jen.Add(x).Dot(expr.Sel.Name), nil

	default:
		return nil, errors.Errorf("unexpected expr type: %T", expr)
	}
}

func (ctx *generateContext) generateFuncWrapper(u *Unit, wf *WrapFunc, cfuncprefix string, altcname func(string) (string, bool), preferC bool, prependCode ...jen.Code) (jen.Code, error) {
	spec, err := u.cfuncSpec(wf.CName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var returnTypes []jen.Code
	if spec.ResultType != nil {
		if !isCVoid(spec.ResultType) {
			typeCode, err := ctx.renderExpr(spec.ResultType, true, nil, false)
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
		} else if typeCode, err = ctx.renderExpr(param.Type, true, nil, false); err != nil {
			return nil, errors.WithStack(err)
		}

		for _, name := range param.Names {
			paramCodes = append(paramCodes, jen.Id(caseconv.CamelCase(name.Name)).Add(typeCode))
		}
	}

	funcStmts := prependCode
	callParams := make([]jen.Code, 0, len(spec.Params))
	var i int
	for _, param := range spec.Params {
		for _, name := range param.Names {
			var paramCode jen.Code = jen.Id(name.Name)
			if override := u.paramType(wf.CName, i); override != nil {
				exprCode, err := ctx.renderExpr(param.Type, true, altcname, preferC)
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

	call := jen.Qual("C", cfuncprefix+wf.CName).Params(callParams...)

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

	// funcStmts = append([]jen.Code{
	// 	jen.Qual("log", "Println").Call(jen.Lit("calling " + wf.CName)),
	// 	jen.Defer().Qual("log", "Println").Call(jen.Lit("called " + wf.CName)),
	// }, funcStmts...)

	return jen.Func().Id(wf.GoName).Params(paramCodes...).Add(returnTypes...).Block(funcStmts...), nil
}

func generateTypes(ctx *generateContext, libname string, u *Unit) error {
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
					} else if typeCode, err = ctx.renderExpr(field.Type, true, nil, false); err != nil {
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

	if !shouldOutput {
		return nil
	}

	var buf bytes.Buffer
	if err := f.Render(&buf); err != nil {
		return errors.WithStack(err)
	}

	if err := ioutil.WriteFile(libname+"_gen.go", buf.Bytes(), 0666); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func generateStaticFunctions(ctx *generateContext, u *Unit) error {
	f := jen.NewFile(ctx.name)
	f.CgoPreamble(u.CgoPreamble)
	f.HeaderComment("Code generated by robots; DO NOT EDIT.")
	f.HeaderComment("+build !av_dynamic")

	var shouldOutput bool
	for _, wf := range u.WrapFuncs {
		code, err := ctx.generateFuncWrapper(u, wf, "", nil, false)
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

	if err := ioutil.WriteFile("static_gen.go", buf.Bytes(), 0666); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func generateDynamicFunctions(ctx *generateContext, libname string, u *Unit) error {
	f := jen.NewFile(ctx.name)
	f.HeaderComment("Code generated by robots; DO NOT EDIT.")
	f.HeaderComment("+build av_dynamic")

	forwardDecls := map[string]bool{}
	var cfuncLines, initBodyLines []string

	initBodyLines = append(initBodyLines, fmt.Sprintf("handle = dlopen(\"lib%s.so\", RTLD_NOW | RTLD_GLOBAL);", libname))

	cfuncprefix := "dyn_"
	for _, wf := range u.WrapFuncs {
		spec, err := u.cfuncSpec(wf.CName)
		if err != nil {
			return errors.WithStack(err)
		}

		cReturnType, ok := ctypeFromGoExpr(spec.ResultType)
		if !ok {
			panic("return type is not a c type?")
		}

		if strings.HasPrefix(cReturnType, "struct ") {
			forwardDecls[strings.TrimRight(cReturnType, " *")+";"] = !strings.ContainsRune(cReturnType, '*')
		}

		var paramTypes, paramDecls, paramIdents []string
		for i, field := range spec.Params {
			ctype, ok := ctypeFromGoExpr(field.Type)
			if !ok {
				log.Printf("%s: %#v", wf.CName, field.Type)
				panic("param type is not a c type?")
			}

			pname := "p" + strconv.Itoa(i)
			if strings.Contains(ctype, "%s") {
				paramTypes = append(paramTypes, fmt.Sprintf(ctype, pname))
				paramDecls = append(paramDecls, fmt.Sprintf(ctype, pname))
			} else {
				paramTypes = append(paramTypes, ctype)
				paramDecls = append(paramDecls, ctype+" "+pname)
			}
			paramIdents = append(paramIdents, pname)

			if strings.HasPrefix(ctype, "struct ") {
				forwardDecls[strings.TrimRight(ctype, " *")+";"] = !strings.ContainsRune(ctype, '*')
			}
		}

		var callPrefix string
		if cReturnType != "void" {
			callPrefix = "return "
		}

		cfuncLines = append(cfuncLines, "static "+cReturnType+" (*_"+wf.CName+")("+strings.Join(paramTypes, ", ")+");")
		cfuncLines = append(cfuncLines, ""+cReturnType+" "+cfuncprefix+wf.CName+"("+strings.Join(paramDecls, ", ")+") {\n    "+callPrefix+"_"+wf.CName+"("+strings.Join(paramIdents, ", ")+");\n};")
		initBodyLines = append(initBodyLines, "_"+wf.CName+" = dlsym(handle, \""+wf.CName+"\");")
	}

	for i, line := range initBodyLines {
		initBodyLines[i] = line + "\n    if (ret = dlerror()) {\n        return ret;\n    }"
	}

	initBodyLines = append(initBodyLines, "return 0;")

	forwardDeclLines := make([]string, 0, len(forwardDecls))
	for forwardDecl, needsCompletion := range forwardDecls {
		if needsCompletion {
			forwardDecl = forwardDecl[:len(forwardDecl)-1] + "{}" + ";"
			// TODO actually implement
		}

		forwardDeclLines = append(forwardDeclLines, forwardDecl)
	}

	sort.Strings(forwardDeclLines)

	loadLibCName := "goav_load_" + libname

	f.CgoPreamble("#cgo LDFLAGS: -ldl\n\n#include <stdlib.h>\n#include <stdint.h>\n#include <dlfcn.h>\n\n" + strings.Join(forwardDeclLines, "\n") + "\n\nstatic void *handle = 0;\n\n" + strings.Join(cfuncLines, "\n\n") + "\n\nchar *" + loadLibCName + "() {\n    char *ret;\n    " + strings.Join(initBodyLines, "\n    ") + "\n}")

	f.Add(jen.Var().Defs(
		jen.Id("initOnce").Qual("sync", "Once"),
		jen.Id("initError").Error(),
		jen.Id("initFuncs").Index().Func().Params(),
	))

	f.Add(jen.Func().Id("dynamicInit").Params().Block(
		jen.Id("initOnce").Dot("Do").Call(
			jen.Func().Params().Block(
				// jen.Qual("log", "Println").Call(jen.Lit("initializing "+libname)),
				jen.If(jen.Id("ret").Op(":=").Qual("C", loadLibCName).Call(), jen.Id("ret").Op("!=").Nil()).Block(
					jen.Id("initError").Op("=").Qual("github.com/pkg/errors", "Errorf").Call(
						jen.Lit("failed to initialize lib"+libname+": %s"),
						jen.Qual("C", "GoString").Call(jen.Id("ret")),
					),
				).Else().Block(
					jen.For(jen.List(jen.Id("_"), jen.Id("f")).Op(":=").Range().Id("initFuncs")).Block(
						jen.Id("f").Call(),
					),
				),
				// jen.Qual("log", "Println").Call(jen.Lit("initialized "+libname)),
			),
		),
		jen.If(jen.Id("initError").Op("!=").Nil()).Block(
			jen.Panic(jen.Id("initError")),
		),
	))

	initCode := jen.Id("dynamicInit").Call()

	var shouldOutput bool
	for _, wf := range u.WrapFuncs {
		code, err := ctx.generateFuncWrapper(u, wf, cfuncprefix, goIntTypeToCIntType, true, initCode)
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

	if err := ioutil.WriteFile("dynamic_gen.go", buf.Bytes(), 0666); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func generate(ctx *generateContext, filename string, u *Unit) error {
	ext := path.Ext(filename)
	libname := filename[:len(filename)-len(ext)]
	if err := generateTypes(ctx, libname, u); err != nil {
		return errors.WithStack(err)
	}

	if err := generateStaticFunctions(ctx, u); err != nil {
		return errors.WithStack(err)
	}

	if err := generateDynamicFunctions(ctx, libname, u); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
