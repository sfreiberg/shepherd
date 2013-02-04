package actions

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type FilePermissions interface {
	GetDestination() string
	GetPermissions() string
	GetUser() string
	GetGroup() string
	GetResult() *Result
}

func UpdatePermissions(fp FilePermissions) error {
	// If nothing is set there's nothing to do so let's finish
	if fp.GetPermissions() == "" && fp.GetUser() == "" && fp.GetGroup() == "" {
		return nil
	}

	perms, err := GetPerms(fp)
	if err != nil {
		return err
	}

	if fp.GetPermissions() != "" && perms.Perms != fp.GetPermissions() {
		cmd := exec.Command("/bin/chmod", fp.GetPermissions(), fp.GetDestination())
		if out, err := cmd.CombinedOutput(); err != nil {
			fp.GetResult().Output = string(out)
			return err
		}
		fp.GetResult().Changed = true
	}

	user := ""
	if fp.GetUser() != "" && fp.GetUser() != perms.UserId && fp.GetUser() != perms.UserName {
		user = fp.GetUser()
	}
	group := ""
	if fp.GetGroup() != "" && fp.GetGroup() != perms.GroupId && fp.GetGroup() != perms.GroupName {
		group = fp.GetGroup()
	}

	if user != "" || group != "" {
		ownership := fmt.Sprintf("%s:%s", user, group)
		cmd := exec.Command("chown", ownership, fp.GetDestination())
		if out, err := cmd.CombinedOutput(); err != nil {
			fp.GetResult().Output = string(out)
			return err
		}
		fp.GetResult().Changed = true
	}

	return nil
}

type Perms struct {
	Perms     string
	UserId    string
	UserName  string
	GroupId   string
	GroupName string
}

func GetPerms(fp FilePermissions) (*Perms, error) {
	// statFormatFlag and statFormat are constants set in platform specific files
	// * perms_darwin.go
	// * perms_linux.go
	cmd := exec.Command("/usr/bin/stat", statFormatFlag, statFormat, fp.GetDestination())
	out, err := cmd.CombinedOutput() //  results are returned as: "perms uid user gid group"
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(out))
	if len(fields) != 5 {
		return nil, errors.New("Incorrect number of fields returned from stat.")
	}

	p := &Perms{
		Perms:     fields[0],
		UserId:    fields[1],
		UserName:  fields[2],
		GroupId:   fields[3],
		GroupName: fields[4],
	}
	return p, nil
}
