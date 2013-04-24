package main

import (
	"github.com/robertkrimen/otto"

	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

var (
	// This is a global javascript interpreter.
	// TODO: Remove this global interpreter because we might want multiples running
	js *otto.Otto
	// Number of minutes between runs
	runInterval int
	// Path to the main shepherd directory which should contain a file named
	// shepherd.js.
	shepherdPath string
	// The current working directory. Used to keep track of which parent
	// directory to use. For example when inside a module we want to work
	// from that modules directory.
	cwd string
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
	js.Set("include", include)
}

func RunClient() {
	flagSet := flag.NewFlagSet("client", flag.ExitOnError)
	flagSet.StringVar(&shepherdPath, "dir", "/etc/shepherd", "Location of shepherd configs")
	flagSet.IntVar(&runInterval, "interval", 30, "Number of minutes between runs")
	if err := flagSet.Parse(os.Args[2:]); err != nil {
		// TODO: Log this before panic and ideally don't panic
		panic(err)
	}
	cwd = shepherdPath

	for {
		go executeMainJS()
		time.Sleep(time.Duration(runInterval) * time.Minute)
	}
}

func RunStandalone() {
	jsFiles := os.Args[1:]

	if len(jsFiles) == 0 {
		fmt.Println("You must provide at least one javascript file.")
		return
	}

	// Read and execute javascript
	executeJSFiles(jsFiles)
}

func executeMainJS() {
	jsFile := path.Join(shepherdPath, "shepherd.js")
	jsFiles := []string{jsFile}
	fmt.Println(jsFiles)
	executeJSFiles(jsFiles)
}

func executeJSFiles(jsFiles []string) {
	initJavascript()

	for _, f := range jsFiles {
		executeJSFile(f)
	}

	fmt.Println("\nFinished.")
}

func executeJSFile(file string) {
	fmt.Printf("Executing %v...\n", file)
	b, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err.Error())
	}
	_, err = js.Run(string(b))
	if err != nil {
		fmt.Println("Error executing: ", file)
		fmt.Println(err)
	}
}
