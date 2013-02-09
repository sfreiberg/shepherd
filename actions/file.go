package actions

import (
	"io"
	"os"
)

type File struct {
	Source      string
	Destination string
	Options     map[string]interface{}
	Result
}

func NewFile(source, destination string) *File {
	f := &File{Source: source, Destination: destination}
	return f
}

func (f *File) Run() {
	// If the file exist, check the hashes
	if f.FileExists() {
		destHash, err := FileHash(f.Destination)
		if err != nil {
			f.Result.Error = err.Error()
			return
		}

		srcHash, err := FileHash(f.Source)
		if err != nil {
			f.Result.Error = err.Error()
			return
		}
		if destHash != srcHash {
			if err := f.Copy(); err != nil {
				f.Result.Error = err.Error()
				return
			}
		}
		// If the file doesn't exist, create it
	} else {
		err := f.Copy()
		if err != nil {
			f.Result.Error = err.Error()
			return
		}
	}

	err := UpdatePermissions(f)
	if err != nil {
		f.Result.Error = err.Error()
		return
	}
	f.Result.Success = true
}

func (f *File) FileExists() bool {
	_, err := os.Stat(f.Destination)
	return !os.IsNotExist(err)
}

func (f *File) Copy() error {
	srcFile, err := os.Open(f.Source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(f.Destination, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	f.Result.Changed = true

	return nil
}

func (f *File) GetDestination() string {
	return f.Destination
}

func (f *File) GetPermissions() string {
	perms := f.Options["perms"]
	if perms != nil {
		return perms.(string)
	}
	return ""
}

func (f *File) GetUser() string {
	user := f.Options["user"]
	if user != nil {
		return user.(string)
	}
	return ""
}

func (f *File) GetGroup() string {
	group := f.Options["group"]
	if group != nil {
		return group.(string)
	}
	return ""
}

func (f *File) GetResult() *Result {
	return &f.Result
}
