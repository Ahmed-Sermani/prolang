package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/Ahmed-Sermani/prolang/interpreter"
	"github.com/Ahmed-Sermani/prolang/parser"
	"github.com/Ahmed-Sermani/prolang/reporting"
	"github.com/Ahmed-Sermani/prolang/scanner"
)

var HadError = false

func main() {
	if len(os.Args) > 2 {
		log.Println("Usage: code [script]")
		os.Exit(64)
	} else if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
		runPrompt()
	}
}

func check(e error) {
	if e != nil {
		log.Panicln(e)
	}
}

func runFile(path string) {
	bytes, err := ioutil.ReadFile(path)
	check(err)
	run(string(bytes))
	if reporting.HadError() {
		os.Exit(65)
	}
}

func runPrompt() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
		}
		run(line)
		// reset the flag in the interactive loop. If the user makes a mistake, it shouldn’t kill their entire session.
		reporting.UnsetError()
	}

}
func run(source string) {
	scanner := scanner.New(source)
	tokens := scanner.ScanTokens()
	p := parser.New(tokens)
	expr, err := p.Parse()
	if err != nil || reporting.HadError() {
		return
	}

	// stop if there is a syntax error
	if reporting.HadError() {
		return
	}
	inter := interpreter.New()
	inter.Interpret(expr)
	// pv := parser.PrintVisitor{}
	// res, _ := pv.Print(expr)
	// fmt.Println(res)
}
