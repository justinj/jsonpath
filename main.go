package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/justinj/jsonpath/jsonpath"
)

func main() {
	flag.Parse()
	program := flag.Args()
	machine, err := jsonpath.NewNaiveEvaler(program[0])
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		var obj interface{}
		json.Unmarshal([]byte(line), &obj)
		result, err := machine.Run(obj)
		if err != nil {
			panic(err)
		}
		if len(result) != 1 {
			panic(fmt.Sprintf("expected single result, got %d", len(result)))
		}
		res, err := json.Marshal(result[0])
		if err != nil {
			panic(err)
		}
		fmt.Println(string(res))
	}

}
