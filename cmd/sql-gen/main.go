package main

import (
	"flag"
	"fmt"
	generator "github.com/joeandaverde/sql-gen-go"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

func findSQLFiles(root string) map[string][]generator.SQLFile {
	tinySQLFiles := make(map[string][]generator.SQLFile)
	absPath, err := filepath.Abs(root)

	if err != nil {
		panic(err)
	}

	sqlFiles, _ := filepath.Glob(absPath + "/**/*.sql")

	for _, path := range sqlFiles {
		relativePath := strings.Replace(path, absPath+string(filepath.Separator), "", 1)
		relativePath = strings.Replace(relativePath, ".sql", "", 1)

		pathParts := strings.Split(relativePath, string(filepath.Separator))

		for i, p := range pathParts {
			pathParts[i] = strcase.ToCamel(p)
		}

		sqlKey := strcase.ToCamel(strings.Join(pathParts, "_"))

		content, err := ioutil.ReadFile(path)

		if err != nil {
			panic(err)
		}

		sqlSource := string(content)
		newSQL, params, err := generator.Query([]byte(sqlSource), generator.DOLLAR, true)
		tinyParams := makeParams(params)

		tinySQL := generator.SQLFile{
			Path:       path,
			Name:       pathParts[1],
			Key:        sqlKey,
			Params:     tinyParams,
			Content:    sqlSource,
			ReboundSQL: newSQL,
		}

		if items, ok := tinySQLFiles[pathParts[0]]; ok {
			tinySQLFiles[pathParts[0]] = append(items, tinySQL)
		} else {
			var thing []generator.SQLFile
			tinySQLFiles[pathParts[0]] = append(thing, tinySQL)
		}
	}

	return tinySQLFiles
}

func makeParams(params []string) map[string]generator.SQLParam {
	result := make(map[string]generator.SQLParam)
	for i, p := range params {
		result[p] = generator.SQLParam{
			Name:  p,
			Index: i + 1,
		}
	}
	return result
}

func realMain() {
	root := flag.String("root", ".", "root path to find sql files")
	out := flag.String("out", "stdout", "output for generated go")
	perms := flag.String("perms", "0644", "permissions for new file")
	flag.Parse()

	var writer io.Writer
	if *out == "stdout" {
		writer = os.Stdout
	} else {
		perm, err := strconv.ParseInt(*perms, 8, 32)
		if err != nil {
			fmt.Println("incorrect permission format, expected octal string")
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

	files := findSQLFiles(*root)
	generatedCode := generator.Go(files)

	if _, err := writer.Write([]byte(generatedCode)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	realMain()
}
