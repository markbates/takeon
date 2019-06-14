package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		// log.Fatal("you must provide a package name")
		args = append(args, "github.com/markbates/oncer")
	}

	pkg := args[0]

	u := fmt.Sprintf("https://%s.git", pkg)

	ipkg := filepath.Join("internal", strings.ReplaceAll(pkg, "/", string(filepath.Separator)))

	os.RemoveAll(ipkg)

	run("git", "clone", u, ipkg)

	os.RemoveAll(filepath.Join(ipkg, ".git"))

	run("go", "mod", "tidy", "-v")

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
