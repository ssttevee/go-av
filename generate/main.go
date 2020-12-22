package main

import (
	"io/ioutil"
	"log"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/ssttevee/go-fmterrors"
)

func start() error {
	const dirname = "."
	list, err := ioutil.ReadDir(dirname)
	if err != nil {
		return errors.WithStack(err)
	}

	files := make(map[string]*Unit)
	for _, item := range list {
		if item.IsDir() || !strings.HasSuffix(item.Name(), ".go") || strings.HasSuffix(item.Name(), "_test.go") || strings.HasSuffix(item.Name(), "_gen.go") {
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
