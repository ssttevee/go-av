package main

import (
	"io/ioutil"
	"log"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/ssttevee/go-fmterrors"
)

func shouldSkipFile(name string) bool {
	switch {
	default:
		return false

	case !strings.HasSuffix(name, ".go"):
	case strings.HasSuffix(name, "_test.go"):
	case strings.HasSuffix(name, "_gen.go"):
	case name == "static.go":
	case name == "dynamic.go":
	}

	return true
}

func start() error {
	const dirname = "."
	list, err := ioutil.ReadDir(dirname)
	if err != nil {
		return errors.WithStack(err)
	}

	files := make(map[string]*Unit)
	for _, item := range list {
		if item.IsDir() || shouldSkipFile(item.Name()) {
			continue
		}

		filename := path.Join(dirname, item.Name())
		f, err := parse(filename)
		if err != nil {
			return errors.WithStack(err)
		}

		if f == nil {
			continue
		}

		files[filename] = f
	}

	var ctx generateContext
	ctx.ct2gt = make(map[string]*ConvertType)
	for _, f := range files {
		if ctx.name == "" {
			ctx.name = f.PackageName
		} else if ctx.name != f.PackageName {
			return errors.Errorf("found multiple package names: %s != %s", ctx.name, f.PackageName)
		}

		for _, convertType := range f.ConvertTypes {
			ctx.ct2gt[convertType.CName] = convertType
		}
	}

	for filename, f := range files {
		if err := generate(&ctx, filename, f); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func main() {
	if err := start(); err != nil {
		log.Println(fmterrors.FormatString(err))
	}
}
