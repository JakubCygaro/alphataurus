package main

import (
	"fmt"
	"go/ast"
	"go/constant"
	// "go/format"
	// "go/parser"
	// "go/token"
	// "go/types"
	// "io/fs"
	"os"
	// "path/filepath"

	// "path/filepath"
	// "strings"
	//
	// "github.com/JakubCygaro/alphataurus/pkg/linker"
	arg "github.com/alexflint/go-arg"
	// "golang.org/x/tools/go/packages"
)

var args struct {
	InputFile  string `arg:"positional,required"`
	TypeName   string `arg:"-t,--type"`
	OutputFile string `arg:"-o,--output"`
}

func exitWithErr(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

type File struct {
	file     *ast.File
	typeName string
	values   []constant.Value
}

func main() {
	if err := arg.Parse(&args); err != nil {
		exitWithErr("%s\n", err.Error())
	}
	// var contents string
	// if f, err := os.ReadFile(args.InputFile); err != nil {
	// 	exitWithErr("%s\n", err.Error())
	// } else {
	// 	contents = string(f)
	// }
	// println(contents)
	// dir := filepath.Dir(args.inputFile)
	// fset := token.NewFileSet()
	// f, err := parser.ParseFile(fset, args.InputFile, nil, 0)
	// if err != nil {
	// 	exitWithErr("%s\n", err.Error())
	// }
	// ast.Print(fset, f)
	// cfg := &packages.Config{
	// 	Mode: packages.NeedName |
	// 		packages.NeedTypes |
	// 		packages.NeedTypesInfo |
	// 		packages.NeedSyntax | packages.NeedFiles,
	// 	Tests:      false,
	// 	BuildFlags: []string{},
	// 	Logf:       nil,
	// }
	// pkgs, err := packages.Load(cfg, fmt.Sprintf("file=%s", args.InputFile))
	// if err != nil {
	// 	exitWithErr("%s\n", err.Error())
	// }
}
