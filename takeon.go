package takeon

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/markbates/takeon/internal/filex"
	"github.com/markbates/takeon/internal/takeon/github.com/fatih/astrewrite"
	"github.com/gobuffalo/here"
)

const intpkg = "internal/takeon"

type Options struct {
	PkgName		string	// github.com/foo/bar
	Undo		bool
	info		here.Info
	intPkgPath	string	// internal/takeon/github.com/foo/bar
	replacePkg	string	// github.com/x/y/internal/takeon/github.com/foo/bar
}

func Me(opts Options) error {
	defer run("go", "mod", "tidy", "-v")
	if len(opts.PkgName) == 0 {
		return fmt.Errorf("you must pass at least one package name")
	}
	hi, err := here.Dir(".")
	if err != nil {
		return err
	}
	opts.info = hi
	opts.intPkgPath = strings.ReplaceAll(filepath.Join(intpkg, opts.PkgName), "/", string(filepath.Separator))
	opts.replacePkg = path.Join(hi.Module.Path, intpkg, opts.PkgName)

	os.RemoveAll(opts.intPkgPath)
	os.MkdirAll(opts.intPkgPath, 0755)

	if err := clone(opts); err != nil {
		return err
	}

	return rewrite(opts)
}

func clone(opts Options) error {
	if opts.Undo {
		return nil
	}
	hi, err := here.Package(opts.PkgName)
	if err != nil {
		return err
	}

	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		input, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		fp := strings.TrimPrefix(path, hi.Dir)
		fp = filepath.Join(opts.intPkgPath, fp)

		os.MkdirAll(filepath.Dir(fp), 0755)
		return ioutil.WriteFile(fp, input, 0644)
	}

	fn = filex.SkipDir(".git", fn)
	fn = filex.SkipDir("node_modules", fn)
	fn = filex.SkipDir("vendor", fn)
	fn = filex.SkipBase("go.mod", fn)
	fn = filex.SkipBase("go.sum", fn)
	fn = filex.SkipSuffix("_test.go", fn)

	return filepath.Walk(hi.Dir, fn)
}

func rewrite(opts Options) error {
	return filepath.Walk(opts.info.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		if strings.Contains(path, opts.intPkgPath) {
			return nil
		}

		if err := rewriteFile(path, opts); err != nil {
			return err
		}

		return nil
	})
}

func rewriteFile(p string, opts Options) error {
	fset := token.NewFileSet()
	src, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}

	f, err := parser.ParseFile(fset, p, src, parser.ParseComments)
	if err != nil {
		return err
	}

	rewritten := astrewrite.Walk(f, func(n ast.Node) (ast.Node, bool) {
		is, ok := n.(*ast.ImportSpec)
		if !ok {
			return n, true
		}
		if is.Path == nil {
			return n, true
		}
		if is.Path.Value == strconv.Quote(opts.PkgName) {
			is.Path.Value = strconv.Quote(opts.replacePkg)
		}
		if opts.Undo {
			if is.Path.Value == strconv.Quote(opts.replacePkg) {
				is.Path.Value = strconv.Quote(opts.PkgName)
			}
		}

		return n, true
	})

	ff, err := os.Create(p)
	if err != nil {
		return err
	}
	defer ff.Close()
	printer.Fprint(ff, fset, rewritten)

	return nil
}

func run(s string, args ...string) error {
	fmt.Println(s, strings.Join(args, " "))
	c := exec.Command(s, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
