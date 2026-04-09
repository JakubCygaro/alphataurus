package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JakubCygaro/alphataurus/pkg/linker"
	arg "github.com/alexflint/go-arg"
)

var args struct {
	Input  []string `arg:"positional,required"`
	Output string `arg:"-o,--output" help:"output file path"`
}

func main() {
	arg.MustParse(&args)
	ld := linker.NewLinker()
	files := make([]linker.LinkerInput, 0, len(args.Input))
	for _, f:= range args.Input {
		files = append(files, linker.SourcePath(f))
	}
	out, err := ld.Link(files)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(-1)
	}
	var outFile *os.File
	if args.Output != "" {
		outFile, err = os.Create(args.Output)
	} else {
		name := "a"
		outFile, err = os.Create(fmt.Sprintf("%s.aelf", strings.TrimSuffix(name, filepath.Ext(name))))
	}
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(-1)
	}
	defer outFile.Close()
	if _, err := outFile.Write(out.Write()); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(-1)
	}
}

