//go:build linux

package packagemanager

import "os/exec"

var pmcommands = []string{
	"apt",
	"dnf",
	"pacman",
	"zypper",
	"nix-env",
}

func Find(osid string) PackageManager {
	for _, pmname := range pmcommands {
		_, err := exec.LookPath(pmname)
		if err == nil {
			return newPackageManager(pmname, osid)
		}
	}
	return nil
}

func newPackageManager(pmname string, osid string) PackageManager {
	switch pmname {
	case "apt":
		return NewApt(osid)
	case "dnf":
		return NewDnf(osid)
	case "pacman":
		return NewPacman(osid)
	case "zypper":
		return NewZypper(osid)
	case "nix-env":
		return NewNixpkgs(osid)
	}
	return nil
}
