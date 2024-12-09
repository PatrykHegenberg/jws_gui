//go:build linux
// +build linux

package packagemanager

import (
	"os/exec"
	"regexp"
	"strings"
)

type Pacman struct {
	name string
	osid string
}

func NewPacman(osid string) *Pacman {
	return &Pacman{
		name: "pacman",
		osid: osid,
	}
}

func (p *Pacman) Packages() packagemap {
	universalPackages := GenerateUniversalPackages()
	aptSpecificPackages := packagemap{}
	for key, value := range aptSpecificPackages {
		universalPackages[key] = value
	}

	return universalPackages
}

func (p *Pacman) InstallCommand(pkg *Package) string {
	if pkg.SystemPackage == false {
		return pkg.NativePackageName[p.osid]
	}
	return "pacman -S " + pkg.NativePackageName[p.name] + " --noconfirm"
}

func (p *Pacman) Name() string {
	return p.name
}

func (p *Pacman) PackageInstalled(pkg *Package) (bool, error) {
	if pkg.SystemPackage == false {
		return false, nil
	}
	stdout, err := exec.Command("pacman", "-Q", pkg.NativePackageName[p.name]).Output()
	if err != nil {
		_, ok := err.(*exec.ExitError)
		if ok {
			return false, nil
		}
		return false, err
	}

	splitoutput := strings.Split(string(stdout), "\n")
	for _, line := range splitoutput {
		if strings.HasPrefix(line, pkg.Name) {
			splitline := strings.Split(line, " ")
			pkg.Version = strings.TrimSpace(splitline[1])
		}
	}

	return true, err
}

func (p *Pacman) PackageAvailable(pkg *Package) (bool, error) {
	if pkg.SystemPackage == false {
		return false, nil
	}
	output, err := exec.Command("pacman", "-Si", pkg.NativePackageName[p.name]).Output()
	if err != nil {
		_, ok := err.(*exec.ExitError)
		if ok {
			return false, nil
		}
		return false, err
	}

	reg := regexp.MustCompile(`.*Version.*?:\s+(.*)`)
	matches := reg.FindStringSubmatch(string(output))
	pkg.Version = ""
	noOfMatches := len(matches)
	if noOfMatches > 1 {
		pkg.Version = strings.TrimSpace(matches[1])
	}

	return true, nil
}
