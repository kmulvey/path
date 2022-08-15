package main

import (
	"flag"
	"fmt"

	"github.com/kmulvey/path"
)

func main() {
	var files path.Path
	flag.Var(&files, "path", "path to files")
	flag.Parse()
	fmt.Println(files)
}
