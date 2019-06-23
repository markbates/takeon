package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/markbates/takeon"
)

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

var verbose bool
var remove bool

func main() {
	//
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

	err := takeon.Me(takeon.Options{
		PkgName:	pkg,
		Undo:		remove,
	})

	if err != nil {
		log.Fatal(err)
	}
}
