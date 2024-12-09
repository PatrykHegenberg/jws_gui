//go:build linux
// +build linux

package packagemanager

import (
	"os/exec"
	"sort"
)

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
func Dependencies(p PackageManager) (DependencyList, error) {

	var dependencies DependencyList

	for name, packages := range p.Packages() {
		dependency := &Dependency{Name: name}
		for _, pkg := range packages {
			dependency.Optional = pkg.Optional
			dependency.External = !pkg.SystemPackage
			dependency.InstallCommand = p.InstallCommand(pkg)
			packageavailable, err := p.PackageAvailable(pkg)
			if err != nil {
				return nil, err
			}
			if packageavailable {
				dependency.Version = pkg.Version
				dependency.PackageName = pkg.Name
				installed, err := p.PackageInstalled(pkg)
				if err != nil {
					return nil, err
				}
				if installed {
					dependency.Installed = true
					dependency.Version = pkg.Version
					if !pkg.SystemPackage {
						dependency.Version = AppVersion(name)
					}
				} else {
					dependency.InstallCommand = p.InstallCommand(pkg)
				}
				break
			}
		}
		dependencies = append(dependencies, dependency)
	}

	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Name < dependencies[j].Name
	})

	return dependencies, nil
}
func AppVersion(name string) string {

	if name == "gcc" {
		return gccVersion()
	}

	if name == "pkg-config" {
		return pkgConfigVersion()
	}

	if name == "npm" {
		return npmVersion()
	}

	if name == "docker" {
		return dockerVersion()
	}

	return ""

}
func gccVersion() string       { return "gcc" }
func pkgConfigVersion() string { return "pkg-config" }
func npmVersion() string       { return "npm" }
func dockerVersion() string    { return "docker" }
