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
	l := assembler.NewLexer(*bufio.NewReader(file))
	for {
		err := l.ReadNextToken()
		if err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(-1)
		}
		if l.CurrentToken().Ty == assembler.TOKEN_TEOF {
			fmt.Println("EOF")
			os.Exit(0)
		}
		fmt.Println(l.CurrentToken())
	}
}
