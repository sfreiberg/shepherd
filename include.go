package main

import (
	"path"
)

func Include(module string) {
	cwd = path.Join(shepherdPath, module)
	file := path.Join(cwd, module+".js")
	executeJSFile(file)
	cwd = path.Join(shepherdPath)
}
