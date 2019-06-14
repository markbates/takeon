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
	"strings"

	"github.com/markbates/takeon/internal/github.com/fatih/astrewrite"
)

var verbose bool
var remove bool

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
	defer run("go", "mod", "tidy", "-v")

	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.BoolVar(&remove, "out", false, "undoes things")
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		log.Fatal("you must pass at least one package name")
	}

	pkg := args[0]

	if pkg == "me" {
		fmt.Println(lyrics)
		return
	}

	ipkg := filepath.Join("internal", strings.ReplaceAll(pkg, "/", string(filepath.Separator)))
	os.RemoveAll(ipkg)

	if remove {
		rewrite(pkg, ipkg)
		return
	}

	clone(pkg, ipkg)

	rewrite(pkg, ipkg)

}

func clone(pkg, ipkg string) {
	gargs := []string{"clone"}
	if verbose {
		gargs = append(gargs, "-v")
	}
	u := fmt.Sprintf("https://%s.git", pkg)
	gargs = append(gargs, u, ipkg)
	run("git", gargs...)

	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)

	os.Chdir(ipkg)
	os.RemoveAll(".git")
	os.RemoveAll("go.mod")
	os.RemoveAll("go.sum")
	// run("go", "mod", "init")
	// run("go", "mod", "tidy", "-v")

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

		ipkg := path.Join(module, "internal", pkg)
		is.Path.Value = strings.ReplaceAll(is.Path.Value, ipkg, pkg)
		if remove {
			return n, true
		}
		is.Path.Value = strings.ReplaceAll(is.Path.Value, pkg, ipkg)

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

const lyrics = `Take On Me
A-ha

We're talking away
I don't know what
I'm to say I'll say it anyway
Today's another day to find you
Shying away
I'll be coming for your love, okay?

Take on me (take on me)
Take me on (take on me)
I'll be gone
In a day or two

So needless to say
I'm odds and ends
But I'll be stumbling away
Slowly learning that life is okay
Say after me
It's no better to be safe than sorry

Take on me (take on me)
Take me on (take on me)
I'll be gone
In a day or two`
