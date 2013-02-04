package facts

import (
	"os"
	"runtime"
	"strings"
)

type Facts struct {
	Hostname string
	Domain   string
	Fqdn     string
	Cpus     int
	Os       string
}

func FindFacts() *Facts {
	f := &Facts{}

	// get the domain info
	if fqdn, err := os.Hostname(); err == nil {
		f.Fqdn = fqdn
		a := strings.SplitN(fqdn, ".", 2)
		if len(a) == 2 {
			f.Hostname = a[0]
			f.Domain = a[1]
		} else if len(a) == 1 {
			f.Hostname = a[0]
		}
	}

	f.Cpus = runtime.NumCPU()

	f.Os = GetOs()

	return f
}
