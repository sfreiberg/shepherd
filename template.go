package main

import (
	"github.com/flosch/pongo"

	"crypto/md5"
	"fmt"
	"io"
	"os"
)

type Template struct {
	Source      string
	Destination string
	Context     map[string]interface{}
	Options     map[string]interface{}
	Result
}

func NewTemplate(source, destination string) *Template {
	t := &Template{Source: source, Destination: destination}
	return t
}

func (t *Template) Run() {
	tmpl, err := t.CreateTemplate()
	if err != nil {
		t.Result.Error = err.Error()
		return
	}

	// If the file exist, check the hashes
	if t.FileExists() {
		fileHash, err := t.FileHash()
		if err != nil {
			t.Result.Error = err.Error()
			return
		}

		if fileHash != t.HashTemplate(tmpl) {
			if err := t.Save(tmpl); err != nil {
				t.Result.Error = err.Error()
				return
			}
		}
		// If the file doesn't exist, create it
	} else {
		err := t.Save(tmpl)
		if err != nil {
			t.Result.Error = err.Error()
			return
		}
	}

	err = UpdatePermissions(t)
	if err != nil {
		t.Result.Error = err.Error()
		return
	}
	t.Result.Success = true
	return
}

func (t *Template) FileHash() (string, error) {
	file, err := os.Open(t.Destination)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	h := fmt.Sprintf("%x", hash.Sum(nil))
	return h, nil
}

func (t *Template) HashTemplate(str string) string {
	hash := md5.New()
	io.WriteString(hash, str)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (t *Template) CreateTemplate() (string, error) {
	tpl, err := pongo.FromFile(t.Source, nil)
	if err != nil {
		return "", err
	}
	ctx := pongo.Context(t.Context)
	out, err := tpl.Execute(&ctx)
	if err != nil {
		return "", err
	}
	return *out, nil
}

func (t *Template) FileExists() bool {
	_, err := os.Stat(t.Destination)
	return !os.IsNotExist(err)
}

func (t *Template) Save(tmpl string) error {
	file, err := os.OpenFile(t.Destination, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(tmpl); err != nil {
		return err
	}
	t.Result.Changed = true

	return nil
}

func (t *Template) GetDestination() string {
	return t.Destination
}

func (t *Template) GetPermissions() string {
	perms := t.Options["perms"]
	if perms != nil {
		return perms.(string)
	}
	return ""
}

func (t *Template) GetUser() string {
	user := t.Options["user"]
	if user != nil {
		return user.(string)
	}
	return ""
}

func (t *Template) GetGroup() string {
	group := t.Options["group"]
	if group != nil {
		return group.(string)
	}
	return ""
}

func (t *Template) GetResult() *Result {
	return &t.Result
}
