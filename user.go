package main

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

type User struct {
	Username string
	HomeDir  string
	Shell    string
	UserId   int64
	GroupId  int64
	Password string
	Groups   []string
	Result
}

func NewUser(username string) *User {
	return &User{Username: username}
}

func (u *User) Run() {
	currentUser, err := u.GetCurrentUser()

	// if both are nil that most likely means the
	// user doesn't exist so let's create it.
	if currentUser == nil && err == nil {
		err := u.CreateUser()
		if err != nil {
			u.Result.Error = err.Error()
			return
		}
		u.Result.Success = true
		return
	}

	err = u.UpdateUser(currentUser)
	if err != nil {
		u.Result.Error = err.Error()
		return
	}

	u.Result.Success = true
	return
}

func (u *User) CreateUser() error {
	cmd := exec.Command("/usr/sbin/useradd")

	if u.HomeDir != "" {
		cmd.Args = append(cmd.Args, "--home", u.HomeDir)
	}

	if u.Shell != "" {
		cmd.Args = append(cmd.Args, "--shell", u.Shell)
	}

	if u.UserId != 0 {
		cmd.Args = append(cmd.Args, "--uid", strconv.FormatInt(u.UserId, 10))
	}

	if u.GroupId != 0 {
		cmd.Args = append(cmd.Args, "--gid", strconv.FormatInt(u.GroupId, 10))
	}

	if u.Password != "" {
		cmd.Args = append(cmd.Args, "--password", u.Password)
	}

	if len(u.Groups) > 0 {
		groups := u.SecondaryGroupsAsString()
		cmd.Args = append(cmd.Args, "--groups", groups)
	}

	cmd.Args = append(cmd.Args, u.Username)

	out, err := cmd.CombinedOutput()
	u.Result.Output = string(out)
	if err != nil {
		return err // We'll capture this error in the calling function
	}

	u.Result.Changed = true
	return nil
}

func (u *User) UpdateUser(curUser *User) error {
	anythingChanged := false
	cmd := exec.Command("/usr/sbin/usermod")

	if u.HomeDir != "" && u.HomeDir != curUser.HomeDir {
		anythingChanged = true
		cmd.Args = append(cmd.Args, "--home", u.HomeDir)
	}

	if u.Shell != "" && u.Shell != curUser.Shell {
		anythingChanged = true
		cmd.Args = append(cmd.Args, "--shell", u.Shell)
	}

	if u.UserId != 0 && u.UserId != curUser.UserId {
		anythingChanged = true
		cmd.Args = append(cmd.Args, "--uid", strconv.FormatInt(u.UserId, 10))
	}

	if u.GroupId != 0 && u.GroupId != curUser.GroupId {
		anythingChanged = true
		cmd.Args = append(cmd.Args, "--gid", strconv.FormatInt(u.GroupId, 10))
	}

	if u.Password != "" && u.Password != curUser.Password {
		anythingChanged = true
		cmd.Args = append(cmd.Args, "--password", u.Password)
	}

	groups := u.SecondaryGroupsAsString()
	if groups != curUser.SecondaryGroupsAsString() {
		anythingChanged = true
		cmd.Args = append(cmd.Args, "--groups", groups)
	}

	if anythingChanged != true {
		return nil
	}

	cmd.Args = append(cmd.Args, u.Username)

	out, err := cmd.CombinedOutput()
	u.Result.Output = string(out)
	if err != nil {
		return err
	}
	u.Result.Changed = true
	return nil
}

func (u *User) GetCurrentUser() (*User, error) {
	// get user info from /etc/passwd
	cmd := exec.Command("/usr/bin/getent", "passwd", u.Username)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// We aren't returning the error because getent returns an error if the
		// user doesn't exist. In this case that's ok.
		return nil, nil
	}

	passwdRow := strings.TrimSpace(string(out))
	passwdFields := strings.Split(passwdRow, ":")
	if len(passwdFields) != 7 {
		return nil, errors.New("Unexpected number of fields in GetCurrentUser")
	}

	// get user info from /etc/shadow
	cmd = exec.Command("/usr/bin/getent", "shadow", u.Username)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	shadowFields := strings.Split(string(out), ":")

	curUser := &User{
		Username: passwdFields[0],
		HomeDir:  passwdFields[5],
		Shell:    passwdFields[6],
		Password: shadowFields[1],
	}

	curUser.UserId, err = strconv.ParseInt(passwdFields[2], 10, 64)
	if err != nil {
		return nil, err
	}
	curUser.GroupId, err = strconv.ParseInt(passwdFields[3], 10, 64)
	if err != nil {
		return nil, err
	}

	return curUser, nil
}

func (u *User) AddSecondaryGroup(group string) {
	u.Groups = append(u.Groups, group)
}

// returns the array of secondary groups as a single comma delimited string
func (u *User) SecondaryGroupsAsString() string {
	return strings.Join(u.Groups, ",")
}
