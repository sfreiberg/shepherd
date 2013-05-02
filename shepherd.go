package main

import (
	"github.com/sfreiberg/facts"

	"fmt"
	"os"
)

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "server":
			// run server
		case "facts":
			ShowFacts()
		case "client":
			RunClient()
		default:
			RunStandalone()
		}
	}
}

func ShowFacts() {
	allFacts, err := facts.FindFacts().ToPrettyJson()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", allFacts)
}
