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
	Output string   `arg:"-o,--output" help:"output file path"`
	Entry  string   `arg:"-e,--entry" help:"specify an entry point file in the format <FILE PATH>[:SYMBOL] "`
}

func main() {
	arg.MustParse(&args)
	ld := linker.NewLinker()
	files := make([]linker.LinkerInput, 0, len(args.Input))
	opts := linker.LinkingOptions{}
	if args.Entry != "" {
		split := strings.Split(args.Entry, ":")
		if len(split) >= 1 {
			opts.EntryFile = split[0]
		}
		if len(split) == 2 {
			opts.EntryLab = strings.Split(args.Entry, ":")[1]
		}
		if len(split) > 2 {
			fmt.Fprintf(os.Stderr,
				"Entry point path parameter in inproper format `%s`", args.Entry)
		}
	}
	for _, f := range args.Input {
		files = append(files, linker.SourcePath(f))
	}
	out, err := ld.Link(files, opts)
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
		outFile, err = os.Create(
			fmt.Sprintf("%s.aelf",
				strings.TrimSuffix(name, filepath.Ext(name))))
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
