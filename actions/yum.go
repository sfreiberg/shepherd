package actions

import (
	"bytes"
	"os/exec"
	"strings"
)

type Yum struct {
	Packages      map[string]*RpmPackage
	StatusUpdated bool
	Groupinstall  bool
	Result
}

type RpmPackage struct {
	Name      string
	Installed bool
	Version   string // not currently used
}

func NewYum() *Yum {
	yum := &Yum{Packages: make(map[string]*RpmPackage)}
	return yum
}

func (y *Yum) AddPackage(name string) {
	pkg := &RpmPackage{Name: name}
	y.Packages[name] = pkg
}

func (y *Yum) AddPackages(names []string) {
	for _, name := range names {
		y.AddPackage(name)
	}
}

func (y *Yum) Run() {
	cont, err := y.PkgsToInstall()
	if err != nil {
		y.Result.Error = err.Error()
		return
	}
	// if cont is false then there were no packages to install
	if cont == false {
		y.Result.Success = true
		return
	}

	var cmd *exec.Cmd
	if y.Groupinstall {
		cmd = exec.Command("/usr/bin/yum", "groupinstall", "-y", "-q")
	} else {
		cmd = exec.Command("/usr/bin/yum", "install", "-y", "-q")
	}

	for _, pkg := range y.Packages {
		if pkg.Installed == false {
			cmd.Args = append(cmd.Args, pkg.Name)
		}
	}

	out, err := cmd.CombinedOutput()
	y.Result.Output = string(out)
	if err != nil {
		y.Result.Error = err.Error()
		return
	}

	y.Result.Changed = true
	y.Result.Success = true

	return
}

func (y *Yum) UpdatePkgStatus() error {
	if y.Groupinstall {
		cmd := exec.Command("/usr/bin/yum", "grouplist")
		for _, pkg := range y.Packages {
			cmd.Args = append(cmd.Args, pkg.Name)
		}

		out, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}

		for _, pkg := range y.Packages {
			if bytes.IndexAny(out, pkg.Name) >= 0 {
				pkg.Installed = true
			}
		}
	} else {
		cmd := exec.Command("/bin/rpm", "-q", "--queryformat", `%{NAME}\t%{VERSION}\n`)
		for _, pkg := range y.Packages {
			cmd.Args = append(cmd.Args, pkg.Name)
		}

		// Ignoring the error because it will error if _any_ package isn't installed.
		out, _ := cmd.CombinedOutput()

		for _, row := range strings.Split(string(out), "\n") {
			// Make sure we only process packages that are installed
			if strings.HasSuffix(row, "not installed") {
				continue // We're ignoring packages that aren't installed
			}
			columns := strings.Split(row, "\t")
			// if there are 2 columns process the row, otherwise skip it
			if len(columns) == 2 {
				pkg := y.Packages[columns[0]]
				if pkg != nil {
					pkg.Version = columns[1]
					pkg.Installed = true
				}
			}
		}
	}

	y.StatusUpdated = true
	return nil
}

func (y *Yum) PkgsToInstall() (bool, error) {
	if y.StatusUpdated == false {
		if err := y.UpdatePkgStatus(); err != nil {
			return false, err
		}
	}

	for _, pkg := range y.Packages {
		if pkg.Installed == false {
			return true, nil // if a pkg isn't installed then it should be so return true
		}
	}
	// if we made it here there is nothing to install
	return false, nil
}
