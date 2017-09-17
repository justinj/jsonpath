package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"

	"github.com/justinj/jsonpath/jsonpath"
)

func main() {
	flag.Parse()
	program := flag.Args()
	fmt.Println(program)
	machine, err := jsonpath.NewNaiveEvaler(program[0])
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("line")
	}

	result, err := machine.Run(nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}
