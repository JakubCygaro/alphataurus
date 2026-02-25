package main

import (
	"bufio"
	"fmt"
	"os"
	"github.com/JakubCygaro/alphataurus/assembler"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		os.Exit(-1)
	}
	file, err := os.Open(args[0])
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(-1)
	}
	p := assembler.NewParser(*bufio.NewReader(file))
	var ok bool
	ok, err = p.ParseNext()
	for ; ok && err == nil ; ok, err = p.ParseNext(){
		fmt.Println(p.CurrentInst())
	}
	if err != nil {
		os.Stderr.WriteString(err.Error())
	}
}
