package main

import (
	"fmt"
	_ "github/markbates/oncer"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/syncx"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		// log.Fatal("you must provide a package name")
		args = append(args, "github.com/gobuffalo/syncx")
	}

	pkg := args[0]

	u := fmt.Sprintf("https://%s.git", pkg)

	ipkg := filepath.Join("internal", strings.ReplaceAll(pkg, "/", string(filepath.Separator)))

	os.RemoveAll(ipkg)

	run("git", "clone", u, ipkg)

	os.RemoveAll(filepath.Join(ipkg, ".git"))

	rewrite(pkg, ipkg)

	run("go", "mod", "tidy", "-v")

}

var processed = &syncx.StringMap{}

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

		return nil
	})
}

func rewriteFile(f string, pkg string) error {
	fset := token.NewFileSet() // positions are relative to fset
	src, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	f, err := parser.ParseExprFrom(fset, "src.go", src, 0)
	if err != nil {
		return err
	}

	// Inspect the AST and print all identifiers and literals.
	ast.Inspect(f, func(n ast.Node) bool {
		var s string
		switch x := n.(type) {
		case *ast.BasicLit:
			s = x.Value
		case *ast.Ident:
			s = x.Name
		}
		if s != "" {
			fmt.Printf("%s:\t%s\n", fset.Position(n.Pos()), s)
		}
		return true
	})
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
