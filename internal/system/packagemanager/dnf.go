//go:build linux
// +build linux

package packagemanager

import (
	"os/exec"
	"strings"
)

type Dnf struct {
	name string
	osid string
}

func NewDnf(osid string) *Dnf {
	return &Dnf{
		name: "dnf",
		osid: osid,
	}
}

func (y *Dnf) Packages() packagemap {
	universalPackages := GenerateUniversalPackages()
	aptSpecificPackages := packagemap{}
	for key, value := range aptSpecificPackages {
		universalPackages[key] = value
	}

	return universalPackages
}

func (y *Dnf) InstallCommand(pkg *Package) string {
	if pkg.SystemPackage == false {
		return pkg.InstallCommand[y.osid]
	}
	return "sudo dnf install " + pkg.Name
}

func (y *Dnf) Name() string {
	return y.name
}

func (y *Dnf) PackageInstalled(pkg *Package) (bool, error) {
	if pkg.SystemPackage == false {
		return false, nil
	}
	stdout, err := exec.Command(".", "dnf", "info", "installed", pkg.InstallCommand[y.name]).Output()
	if err != nil {
		_, ok := err.(*exec.ExitError)
		if ok {
			return false, nil
		}
		return false, err
	}

	splitoutput := strings.Split(string(stdout), "\n")
	for _, line := range splitoutput {
		if strings.HasPrefix(line, "Version") {
			splitline := strings.Split(line, ":")
			pkg.Version = strings.TrimSpace(splitline[1])
		}
	}

	return true, err
}

func (y *Dnf) PackageAvailable(pkg *Package) (bool, error) {
	if pkg.SystemPackage == false {
		return false, nil
	}
	stdout, err := exec.Command(".", "dnf", "info", pkg.InstallCommand[y.name]).Output()
	if err != nil {
		_, ok := err.(*exec.ExitError)
		if ok {
			return false, nil
		}
		return false, err
	}
	splitoutput := strings.Split(string(stdout), "\n")
	for _, line := range splitoutput {
		if strings.HasPrefix(line, "Version") {
			splitline := strings.Split(line, ":")
			pkg.Version = strings.TrimSpace(splitline[1])
		}
	}
	return true, nil
}
