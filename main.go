package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/markbates/takeon/internal/github.com/fatih/astrewrite"
)

var verbose bool

var module = func() string {
	c := exec.Command("go", "env", "GOMOD")
	b, err := c.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	p := string(bytes.TrimSpace(b))

	f, err := os.Open(p)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	br := bufio.NewReader(f)
	line, _, err := br.ReadLine()
	if err != nil {
		log.Fatal(err)
	}

	pre := []byte("module ")
	if !bytes.HasPrefix(line, pre) {
		log.Fatal("you need a module, sorry")
	}

	res := bytes.TrimPrefix(line, pre)
	res = bytes.TrimSpace(res)

	return string(res)
}()

func main() {

	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.Parse()

	args := flag.Args()

	fmt.Printf("### main.go:41 module (%T) -> %q %+v\n", module, module, module)

	if len(args) == 0 {

		args = append(args, "github.com/fatih/astrewrite")
	}

	pkg := args[0]

	u := fmt.Sprintf("https://%s.git", pkg)

	ipkg := filepath.Join("internal", strings.ReplaceAll(pkg, "/", string(filepath.Separator)))

	os.RemoveAll(ipkg)

	gargs := []string{"clone"}
	if verbose {
		gargs = append(gargs, "-v")
	}
	gargs = append(gargs, u, ipkg)
	run("git", gargs...)

	os.RemoveAll(filepath.Join(ipkg, ".git"))

	rewrite(pkg, ipkg)

	run("go", "mod", "tidy", "-v")

}

func rewrite(pkg string, ipkg string) {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		if strings.Contains(path, ipkg) {
			return nil
		}
		if err := rewriteFile(path, pkg); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func rewriteFile(p string, pkg string) error {
	fset := token.NewFileSet()
	src, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}

	f, err := parser.ParseFile(fset, p, src, 0)
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

		if is.Path.Value != strconv.Quote(pkg) {
			return n, true
		}

		is.Path.Value = strconv.Quote(path.Join(module, "internal", pkg))

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

func run(s string, args ...string) {
	fmt.Println(s, strings.Join(args, " "))
	c := exec.Command(s, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}
