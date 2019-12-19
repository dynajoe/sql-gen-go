package main

import (
	"flag"
	"fmt"
	"github.com/joeandaverde/sql-gen-go"
	"io"
	"os"
	"strconv"
)

func runGenerator() {
	root := flag.String("root", ".", "root path to recursively find sql files")
	pkgName := flag.String("package", "sql", "package name for generated go file")
	out := flag.String("out", "stdout", "output for generated go")
	perms := flag.String("perms", "0644", "permissions for new file")
	flag.Parse()

	var writer io.Writer
	if *out == "stdout" {
		writer = os.Stdout
	} else {
		perm, err := strconv.ParseInt(*perms, 8, 32)
		if err != nil {
			fmt.Println("invalid permission format: expected octal string")
			os.Exit(1)
		}

		file, err := os.OpenFile(*out, os.O_WRONLY, os.FileMode(perm))
		if err != nil {
			fmt.Println("unable to open file for writing")
			os.Exit(1)
		}

		defer file.Close()
		writer = file
	}

	err := generator.Run(*pkgName, *root, writer)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	runGenerator()
}
