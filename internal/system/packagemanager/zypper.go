//go:build linux
// +build linux

package packagemanager

import (
	"os/exec"
	"regexp"
	"strings"
)

type Zypper struct {
	name string
	osid string
}

func NewZypper(osid string) *Zypper {
	return &Zypper{
		name: "zypper",
		osid: osid,
	}
}

func (z *Zypper) Packages() packagemap {
	universalPackages := GenerateUniversalPackages()
	aptSpecificPackages := packagemap{}
	for key, value := range aptSpecificPackages {
		universalPackages[key] = value
	}

	return universalPackages
}

func (z *Zypper) InstallCommand(pkg *Package) string {
	if pkg.SystemPackage == false {
		return pkg.NativePackageName[z.osid]
	}
	return "zypper in " + pkg.NativePackageName[z.name] + " -y"
}

func (z *Zypper) Name() string {
	return z.name
}

func (z *Zypper) PackageInstalled(pkg *Package) (bool, error) {
	if pkg.SystemPackage == false {
		return false, nil
	}
	cmd := exec.Command(".", "zypper", "info", pkg.NativePackageName[z.name])
	cmd.Env = []string{"LANGUAGE", "en_US.utf-8"}
	stdout, err := cmd.Output()
	if err != nil {
		_, ok := err.(*exec.ExitError)
		if ok {
			return false, nil
		}
		return false, err
	}
	reg := regexp.MustCompile(`.*Installed\s*:\s*(Yes)\s*`)
	matches := reg.FindStringSubmatch(string(stdout))
	pkg.Version = ""
	noOfMatches := len(matches)
	if noOfMatches > 1 {
		z.getPackageVersion(pkg, string(stdout))
	}
	return noOfMatches > 1, err
}

func (z *Zypper) PackageAvailable(pkg *Package) (bool, error) {
	if pkg.SystemPackage == false {
		return false, nil
	}
	env := []string{"LANGUAGE", "en_US.utf-8"}
	cmd := exec.Command(".", "zypper", "info", pkg.NativePackageName[z.name])
	cmd.Env = env
	stdout, err := cmd.Output()
	if err != nil {
		_, ok := err.(*exec.ExitError)
		if ok {
			return false, nil
		}
		return false, err
	}

	available := strings.Contains(string(stdout), "Information for package")
	if available {
		z.getPackageVersion(pkg, string(stdout))
	}

	return available, nil
}

func (z *Zypper) getPackageVersion(pkg *Package, output string) {
	reg := regexp.MustCompile(`.*Version.*:(.*)`)
	matches := reg.FindStringSubmatch(output)
	pkg.Version = ""
	noOfMatches := len(matches)
	if noOfMatches > 1 {
		pkg.Version = strings.TrimSpace(matches[1])
	}
}
