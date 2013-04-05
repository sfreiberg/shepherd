package main

import (
	"os"
)

type Directory struct {
	Destination string
	Options     map[string]interface{}
	Result
}

func NewDirectory(path string) *Directory {
	return &Directory{Destination: path}
}

func (d *Directory) Run() {
	// Does the directory exist?
	fileInfo, err := os.Stat(d.Destination)
	if err != nil { // Directory doesn't exist so let's attempt to create it.
		d.CreateDirectory()
		return
	}

	// Make sure it's actually a directory
	if fileInfo.IsDir() == false {
		if err = os.Remove(d.Destination); err != nil {
			d.Result.Error = err.Error()
			return
		} else {
			d.CreateDirectory()
			d.Result.Success = true
			return
		}
	}

	err = UpdatePermissions(d)
	if err != nil {
		d.Result.Error = err.Error()
		return
	}
	d.Result.Success = true
}

func (d *Directory) CreateDirectory() {
	// Create the directory

	// We have to pass something into MkdirAll for perms
	// but the umask may override this so it isn't reliable.
	if err := os.MkdirAll(d.Destination, 0750); err != nil {
		d.Result.Error = err.Error()
		return
	}
	d.Result.Changed = true

	UpdatePermissions(d)
}

func (d *Directory) GetDestination() string {
	return d.Destination
}

func (d *Directory) GetPermissions() string {
	perms := d.Options["perms"]
	if perms != nil {
		return perms.(string)
	}
	return ""
}

func (d *Directory) GetUser() string {
	user := d.Options["user"]
	if user != nil {
		return user.(string)
	}
	return ""
}

func (d *Directory) GetGroup() string {
	group := d.Options["group"]
	if group != nil {
		return group.(string)
	}
	return ""
}

func (d *Directory) GetResult() *Result {
	return &d.Result
}
