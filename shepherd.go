package main

import (
	"os"
)

func main() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "server":
			// run server
		case "client":
			// run client
		default:
			RunStandalone()
		}
	}
}
