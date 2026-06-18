package main

import (
	"fmt"
	"os"

	aelf "github.com/JakubCygaro/alphataurus/pkg/vm/aelf"
	arg "github.com/alexflint/go-arg"
)

var args struct {
	File string `arg:"positional,required"`
}

func main() {
	arg.MustParse(&args)
	if b, err := os.ReadFile(args.File); err != nil {
		fmt.Fprintf(os.Stderr, "Could not read file `%s`\n", args.File)
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	} else if a, err := aelf.LoadAELF(b); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	} else {
		fmt.Fprintf(
			os.Stdout,
			"ALPHATAURUS EXECUTABLE FILE\n"+
				"version: %b.%b.%b.%b\n"+
				"header size: %v bytes\n"+
				"has entry: %v\n"+
				"entry at: 0x%x\n"+
				"code section size: %v bytes\n"+
				"code section at: %v\n"+
				"sdata section size: %v bytes\n"+
				"sdata section at: %v\n"+
				"symbol section size: %v bytes\n"+
				"symbol section at: %v\n"+
				"rels section size: %v bytes\n"+
				"rels section at: %v\n"+
				"",
			byte(a.Version>>24), byte(a.Version>>16),
			byte(a.Version>>8), byte(a.Version),
			a.HeaderSize,
			a.HasEntry,
			a.Entry,
			a.CodeSize,
			a.CodeStart,
			a.StaticDataSize,
			a.StaticDataStart,
			a.SymbolsSize,
			a.StaticDataStart,
			a.RelocsSize,
			a.RelocsStart,
		)
	}
}

