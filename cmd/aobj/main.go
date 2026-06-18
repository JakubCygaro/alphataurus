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
	File string `arg:"positional,required"`
}

func main() {
	arg.MustParse(&args)
	if b, err := os.ReadFile(args.File); err != nil {
		fmt.Fprintf(os.Stderr, "Could not read file `%s`\n", args.File)
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	} else if obj, err := aobj.LoadObjFileHeader(
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
			byte(obj.Version>>24), byte(obj.Version>>16),
			byte(obj.Version>>8), byte(obj.Version),
			obj.HeaderSize,
			obj.HasEntry,
			obj.Entry,
			obj.CodeSize,
			obj.CodeStart,
			obj.StaticDataSize,
			obj.StaticDataStart,
			obj.SymbolsSize,
			obj.StaticDataStart,
			obj.RelocsSize,
			obj.RelocsStart,
		)
	}
}
