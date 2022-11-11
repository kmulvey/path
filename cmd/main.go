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

	fmt.Println("user input: ", files.GivenInput)
	fmt.Println("computed abs path: ", files.ComputedPath.AbsolutePath)
	fmt.Println("number of files found within this path: ", len(files.Files))
}
