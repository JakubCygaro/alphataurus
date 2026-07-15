package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	// "path/filepath"
	// "strings"

	aobj "github.com/JakubCygaro/alphataurus/pkg/vm/obj"
	arg "github.com/alexflint/go-arg"
)

var args struct {
	File        string `arg:"positional,required"`
	ShowSymbols bool   `arg:"-s,--list-symbols" help:"list all symbols in the file"`
}

func showSymbols(header aobj.ObjFileHeader, obj []byte) error {
	syms, err := aobj.LoadSymbolsWithHeader(header, obj)
	if err != nil {
		return err
	}
	fmt.Fprintf(
		os.Stdout,
		"\n=== SYMBOL TABLE === \n",
	)
	for i, s := range syms.InOrder {
		fmt.Fprintf(
			os.Stdout,
			"[%d] %s\n",
			i,
			s.String(),
		)
	}
	return nil
}

func main() {
	arg.MustParse(&args)
	if b, err := os.ReadFile(args.File); err != nil {
		fmt.Fprintf(os.Stderr, "Could not read file `%s`\n", args.File)
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	} else if objHeader, err := aobj.LoadObjFileHeader(
		bufio.NewReader(bytes.NewReader(b))); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	} else {
		fmt.Fprintf(
			os.Stdout,
			"ALPHATAURUS OBJECT FILE\n"+
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
			byte(objHeader.Version>>24), byte(objHeader.Version>>16),
			byte(objHeader.Version>>8), byte(objHeader.Version),
			objHeader.HeaderSize,
			objHeader.HasEntry,
			objHeader.Entry,
			objHeader.CodeSize,
			objHeader.CodeStart,
			objHeader.StaticDataSize,
			objHeader.StaticDataStart,
			objHeader.SymbolsSize,
			objHeader.StaticDataStart,
			objHeader.RelocsSize,
			objHeader.RelocsStart,
		)
		if args.ShowSymbols {
			if err := showSymbols(objHeader, b); err != nil {
				os.Stderr.WriteString(err.Error())
				os.Stderr.WriteString("\n")
				os.Exit(1)
			}
		}
	}
}
