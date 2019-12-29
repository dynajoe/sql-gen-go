package main

import (
	"flag"
	"fmt"
	"github.com/joeandaverde/sql-gen-go"
	"io"
	"io/ioutil"
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

		tempFile, err := ioutil.TempFile(os.TempDir(), "sql-gen")

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer func() {
			_ = tempFile.Close()
			_ = tempFile.Chmod(os.FileMode(perm))
			_ = os.Rename(tempFile.Name(), *out)
		}()

		writer = tempFile
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
