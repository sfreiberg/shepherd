package actions

import (
	"os"
	"os/exec"
	"strings"
)

type Apt struct {
	Packages      map[string]*AptPackage
	StatusUpdated bool // Has UpdatePkgStatus() been run?
	Result
}

type AptPackage struct {
	Name      string
	Installed bool
	Version   string
}

func NewApt() *Apt {
	apt := &Apt{
		Packages:      make(map[string]*AptPackage),
		StatusUpdated: false,
	}

	return apt
}

func (a *Apt) AddPackage(name string) {
	pkg := &AptPackage{Name: name}
	a.Packages[name] = pkg
}

func (a *Apt) AddPackages(names []string) {
	for _, name := range names {
		a.AddPackage(name)
	}
}

func (a *Apt) Run() {
	cont, err := a.PkgsToInstall()
	if err != nil {
		a.Result.Error = err.Error()
		return
	}
	// if cont is false then there were no packages to install
	if cont == false {
		a.Result.Success = true
		return
	}

	// We need to make sure we don't run in interactive mode or
	// modules like mysql will hang.
	os.Setenv("DEBIAN_FRONTEND", "noninteractive")

	cmd := exec.Command("/usr/bin/apt-get", "install", "-y", "-q")
	for _, pkg := range a.Packages {
		if pkg.Installed == false {
			cmd.Args = append(cmd.Args, pkg.Name)
		}
	}

	out, err := cmd.CombinedOutput()
	a.Result.Output = string(out)
	if err != nil {
		a.Result.Error = err.Error()
		return
	}
	a.Result.Success = true
	a.Result.Changed = true
	return
}

func (a *Apt) UpdatePkgStatus() error {
	cmd := exec.Command("/usr/bin/dpkg-query", "-W", "--showformat", `${Package}\t${Status}\t${Version}\n`)
	for _, pkg := range a.Packages {
		cmd.Args = append(cmd.Args, pkg.Name)
	}

	// Ignoring the error because it will error if _any_ package isn't installed.
	out, _ := cmd.CombinedOutput()

	for _, row := range strings.Split(string(out), "\n") {
		columns := strings.Split(row, "\t")
		// if there are 3 columns process the row, otherwise we just skip it for now
		if len(columns) == 3 {
			pkg := a.Packages[columns[0]]
			if pkg != nil {
				pkg.Version = columns[2]
				if columns[1] == "install ok installed" {
					pkg.Installed = true
				} else {
					pkg.Installed = false
				}
			}
		}
	}
	a.StatusUpdated = true
	return nil
}

func (a *Apt) PkgsToInstall() (bool, error) {
	if a.StatusUpdated == false {
		if err := a.UpdatePkgStatus(); err != nil {
			return false, err
		}
	}

	for _, pkg := range a.Packages {
		if pkg.Installed == false {
			return true, nil // if a pkg isn't installed then it should be so return true
		}
	}
	// if we made it here there is nothing to install
	return false, nil
}
