package main

import (
	"github.com/robertkrimen/otto"

	"fmt"
	"io/ioutil"
	"os"
)

var (
	js *otto.Otto
)

func initJavascript() {
	js = otto.New()

	f := FindFacts()
	if jsFacts, err := js.Object(`facts = {}`); err == nil {
		jsFacts.Set("hostname", f.Hostname)
		jsFacts.Set("domain", f.Domain)
		jsFacts.Set("fqdn", f.Fqdn)
		jsFacts.Set("cpus", f.Cpus)
		jsFacts.Set("os", f.Os)
	}

	// Create javascript functions
	js.Set("directory", directory)
	js.Set("symlink", symlink)
	js.Set("template", template)
	js.Set("apt", apt)
	js.Set("yum", yum)
	js.Set("command", command)
	js.Set("user", user)
	js.Set("pg_user", pgUser)
	js.Set("pg_database", pgDatabase)
	js.Set("mysql_user", mysqlUser)
	js.Set("mysql_database", mysqlDatabase)
	js.Set("sleep", sleep)
	js.Set("upstart", upstart)
	js.Set("file", file)
}

func RunStandalone() {
	initJavascript()

	jsFiles := os.Args[1:]

	if len(jsFiles) == 0 {
		fmt.Println("You must provide at least one javascript file.")
		return
	}

	// Read and execute javascript
	for _, f := range jsFiles {
		fmt.Printf("Executing %v...\n", f)
		b, err := ioutil.ReadFile(f)
		if err != nil {
			panic(err.Error())
		}
		_, err = js.Run(string(b))
		if err != nil {
			fmt.Println("Error executing: ", f)
			fmt.Println(err)
		}
	}

	fmt.Println("\nFinished.")
}
