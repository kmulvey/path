package main

import (
	"flag"
	"fmt"

	"github.com/kmulvey/path"
)

func main() {
	var files path.Entry
	flag.Var(&files, "path", "path to files")
	flag.Parse()

	fmt.Println("user input: ", files.AbsolutePath)
	fmt.Println("number of files found within this path: ", len(files.Children))
}
