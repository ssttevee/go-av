package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/pkg/errors"
)

const directiveKeyword = "+gen "

type GoRef struct {
	Pointers  int
	GoPackage string
	GoName    string
}

func parseGoRef(s string) GoRef {
	var pointers int
	for s[0] == '*' {
		pointers++
		s = s[1:]
	}

	if pos := strings.LastIndex(s, "."); pos >= 0 {
		return GoRef{
			Pointers:  pointers,
			GoPackage: s[:pos],
			GoName:    s[pos+1:],
		}
	}

	return GoRef{
		Pointers: pointers,
		GoName:   s,
	}
}

func parseGoRefDefault(s, def string) GoRef {
	ref := parseGoRef(s)
	if ref.GoName == "" {
		ref.GoName = def
	}

	return ref
}

func (t GoRef) Code() jen.Code {
	stars := starsCodes(t.Pointers)
	if t.GoPackage == "" {
		return jen.Add(stars...).Id(t.GoName)
	}

	return jen.Add(stars...).Qual(t.GoPackage, t.GoName)
}

type ConvertType struct {
	CName string
	GoRef
}

type WrapFunc struct {
	CName  string
	GoName string
}

type Unit struct {
	PackageName  string
	CgoPreamble  string
	ConvertTypes []*ConvertType
	WrapFuncs    []*WrapFunc

	FieldTypes map[string]map[string]GoRef
	ParamTypes map[string]map[int]GoRef

	cachedCgoAST *ast.File
}

func parse(filename string) (*Unit, error) {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var fset token.FileSet
	t, err := parser.ParseFile(&fset, path.Base(filename), src, parser.ParseComments)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing %s", path.Base(filename))
	}

	var preamble string
	for _, decl := range t.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range d.Specs {
			s, ok := spec.(*ast.ImportSpec)
			if !ok || s.Path.Value != `"C"` {
				continue
			}

			cg := s.Doc
			if cg == nil && len(d.Specs) == 1 {
				cg = d.Doc
			}

			if cg != nil {
				preamble = cg.Text()
			}
		}
	}

	if preamble == "" {
		//log.Printf("no c preamble found in %s", filename)
		return nil, nil
	}

	convs := []*ConvertType{
		{CName: "int", GoRef: GoRef{GoName: "int32"}},
		{CName: "int8_t", GoRef: GoRef{GoName: "int8"}},
		{CName: "int64_t", GoRef: GoRef{GoName: "int64"}},
		{CName: "uint", GoRef: GoRef{GoName: "uint32"}},
		{CName: "uint8_t", GoRef: GoRef{GoName: "uint8"}},
		{CName: "uint64_t", GoRef: GoRef{GoName: "uint64"}},
		{CName: "uchar", GoRef: GoRef{GoName: "byte"}},
		{CName: "size_t", GoRef: GoRef{GoName: "uint64"}},
		{CName: "double", GoRef: GoRef{GoName: "float64"}},
		{CName: "char", GoRef: GoRef{GoPackage: "github.com/ssttevee/go-av/internal/common", GoName: "CChar"}},
	}
	var funcs []*WrapFunc
	fields := make(map[string]map[string]GoRef)
	params := make(map[string]map[int]GoRef)
	for _, g := range t.Comments {
		for _, c := range g.List {
			index := strings.Index(c.Text, directiveKeyword)
			if index < 0 {
				continue
			}

			directive := c.Text[index+len(directiveKeyword):]

			parts := strings.Split(directive, " ")
			var part1, part2, part3 string
			if len(parts) > 1 {
				part1 = parts[1]
			}

			if len(parts) > 2 {
				part2 = parts[2]
			}

			if len(parts) > 3 {
				part3 = parts[3]
			}

			switch parts[0] {
			case "convtype":
				convs = append(convs, &ConvertType{
					CName: part1,
					GoRef: parseGoRefDefault(part2, part1),
				})

			case "wrapfunc":
				if part2 == "" {
					part2 = part1
				}

				funcs = append(funcs, &WrapFunc{
					CName:  part1,
					GoName: part2,
				})

			case "fieldtype":
				m, ok := fields[part1]
				if !ok {
					m = make(map[string]GoRef)
					fields[part1] = m
				}

				m[part2] = parseGoRef(part3)

			case "paramtype":
				i, err := strconv.ParseInt(part2, 10, 64)
				if err != nil {
					return nil, errors.Errorf("invalid index in directive: %s", directive)
				}

				m, ok := params[part1]
				if !ok {
					m = make(map[int]GoRef)
					params[part1] = m
				}

				m[int(i)] = parseGoRef(part3)

			default:
				log.Printf("unexpected directive: %q", directive)
			}
		}
	}

	if len(convs) == 0 && len(funcs) == 0 {
		return nil, nil
	}

	sort.Slice(convs, func(i, j int) bool {
		return strings.Compare(convs[i].GoName, convs[j].GoName) < 0
	})

	sort.Slice(funcs, func(i, j int) bool {
		return strings.Compare(funcs[i].GoName, funcs[j].GoName) < 0
	})

	return &Unit{
		CgoPreamble:  preamble,
		PackageName:  t.Name.Name,
		ConvertTypes: convs,
		WrapFuncs:    funcs,
		FieldTypes:   fields,
		ParamTypes:   params,
	}, nil
}
