//go:build linux
// +build linux

package packagemanager

import (
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Apt struct {
	name string
	osid string
}

func NewApt(osid string) *Apt {
	return &Apt{
		name: "apt",
		osid: osid,
	}
}

func (a *Apt) Packages() packagemap {
	universalPackages := GenerateUniversalPackages()
	aptSpecificPackages := packagemap{}
	for key, value := range aptSpecificPackages {
		universalPackages[key] = value
	}

	return universalPackages
}

func (a *Apt) InstallCommand(pkg *Package) string {
	if pkg.SystemPackage == false {
		return pkg.NativePackageName[a.osid]
	}
	return "apt install " + pkg.NativePackageName[a.name] + " -y"
}

func (a *Apt) Name() string {
	return a.name
}

func (a *Apt) PackageInstalled(pkg *Package) (bool, error) {
	if pkg.SystemPackage == false {
		return false, nil
	}
	cmd := exec.Command("apt", "list", "-qq", pkg.NativePackageName[a.name])
	var stdo, stde bytes.Buffer
	cmd.Stdout = &stdo
	cmd.Stderr = &stde
	cmd.Env = append(os.Environ(), "LANGUAGE=en")
	err := cmd.Run()
	return strings.Contains(stdo.String(), "[installed]"), err
}

func (a *Apt) PackageAvailable(pkg *Package) (bool, error) {
	if pkg.SystemPackage == false {
		return false, nil
	}
	stdout, err := exec.Command(".", "apt", "list", "-qq", pkg.NativePackageName[a.name]).Output()
	output := a.removeEscapeSequences(string(stdout))
	installed := strings.HasPrefix(output, pkg.Name)
	a.getPackageVersion(pkg, output)
	return installed, err
}

func (a *Apt) removeEscapeSequences(in string) string {
	escapechars, _ := regexp.Compile(`\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])`)
	return escapechars.ReplaceAllString(in, "")
}

func (a *Apt) getPackageVersion(pkg *Package, output string) {
	splitOutput := strings.Split(output, " ")
	if len(splitOutput) > 1 {
		pkg.Version = splitOutput[1]
	}
}
