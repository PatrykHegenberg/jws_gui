//go:build linux
// +build linux

package packagemanager

import (
	"encoding/json"
	"os/exec"
)

type Nixpkgs struct {
	name string
	osid string
}

type NixPackageDetail struct {
	Name    string
	Pname   string
	Version string
}

var available map[string]NixPackageDetail

func NewNixpkgs(osid string) *Nixpkgs {
	available = map[string]NixPackageDetail{}

	return &Nixpkgs{
		name: "nixpkgs",
		osid: osid,
	}
}

func (n *Nixpkgs) Packages() packagemap {
	universalPackages := GenerateUniversalPackages()
	aptSpecificPackages := packagemap{}
	for key, value := range aptSpecificPackages {
		universalPackages[key] = value
	}

	return universalPackages
}

func (n *Nixpkgs) InstallCommand(pkg *Package) string {
	if pkg.SystemPackage == false {
		return pkg.InstallCommand[n.osid]
	}
	return "nix-env -iA " + pkg.Name
}

func (n *Nixpkgs) Name() string {
	return n.name
}

func (n *Nixpkgs) PackageInstalled(pkg *Package) (bool, error) {
	if pkg.SystemPackage == false {
		return false, nil
	}

	stdout, err := exec.Command(".", "nix-env", "--json", "-qA", pkg.InstallCommand[n.name]).Output()
	if err != nil {
		return false, nil
	}

	var attributes map[string]NixPackageDetail
	err = json.Unmarshal([]byte(stdout), &attributes)
	if err != nil {
		return false, err
	}

	installed := false
	for attribute, detail := range attributes {
		if attribute == pkg.Name {
			installed = true
			pkg.Version = detail.Version
		}
		break
	}

	detail, ok := available[pkg.Name]
	if !installed && n.osid == "nixos" && ok {
		cmd := "nix-store --query --requisites /run/current-system | cut -d- -f2- | sort | uniq | grep '^" + detail.Pname + "'"

		if pkg.Library {
			cmd += " | grep 'dev$'"
		}

		stdout, err = exec.Command(".", "sh", "-c", cmd).Output()
		if err != nil {
			return false, nil
		}

		if len(string(stdout)) > 0 {
			installed = true
		}
	}

	return installed, nil
}

func (n *Nixpkgs) PackageAvailable(pkg *Package) (bool, error) {
	if pkg.SystemPackage == false {
		return false, nil
	}

	stdout, err := exec.Command(".", "nix-env", "--json", "-qaA", pkg.InstallCommand[n.name]).Output()
	if err != nil {
		return false, nil
	}

	var attributes map[string]NixPackageDetail
	err = json.Unmarshal(stdout, &attributes)
	if err != nil {
		return false, err
	}

	for attribute, detail := range attributes {
		pkg.Version = detail.Version
		available[attribute] = detail
		break
	}

	return len(pkg.Version) > 0, nil
}
