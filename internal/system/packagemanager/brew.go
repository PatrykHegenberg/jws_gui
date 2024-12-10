//go:build darwin
// +build darwin

package packagemanager

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Homebrew struct {
	name string
	osid string
}

func NewHomebrew(osid string) *Homebrew {
	return &Homebrew{
		name: "brew",
		osid: osid,
	}
}

func (h *Homebrew) Packages() packagemap {
	universalPackages := GenerateUniversalPackages()

	macSpecificPackages := packagemap{
		"openjdk": {
			{
				Name:          "openjdk",
				SystemPackage: true,
				NativePackageName: map[string]string{
					"brew": "openjdk@17",
				},
			},
		},
		"vscode": {
			{
				Name:          "vscode",
				SystemPackage: true,
				NativePackageName: map[string]string{
					"brew": "visual-studio-code",
				},
			},
		},
	}

	for key, value := range macSpecificPackages {
		universalPackages[key] = value
	}

	return universalPackages
}

func (h *Homebrew) InstallCommand(pkg *Package) string {
	packageName := pkg.NativePackageName[h.name]

	if packageName == "" {
		packageName = pkg.NativePackageName[h.osid]
	}

	return fmt.Sprintf("brew install %s", packageName)
}

func (h *Homebrew) Name() string {
	return h.name
}

func (h *Homebrew) PackageInstalled(pkg *Package) (bool, error) {
	if !pkg.SystemPackage {
		return false, nil
	}

	packageName := pkg.NativePackageName[h.name]
	if packageName == "" {
		return false, fmt.Errorf("no Homebrew package name found for %s", pkg.Name)
	}

	cmd := exec.Command("brew", "list")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return strings.Contains(string(output), packageName), nil
}

func (h *Homebrew) PackageAvailable(pkg *Package) (bool, error) {
	if !pkg.SystemPackage {
		return false, nil
	}

	packageName := pkg.NativePackageName[h.name]
	if packageName == "" {
		return false, fmt.Errorf("no Homebrew package name found for %s", pkg.Name)
	}

	cmd := exec.Command("brew", "search", packageName)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	reg := regexp.MustCompile(packageName + `\s+(\d+[\.\d+]*)`)
	matches := reg.FindStringSubmatch(string(output))

	if len(matches) > 1 {
		pkg.Version = matches[1]
		return true, nil
	}

	return false, nil
}

func (h *Homebrew) EnsureInstalled() error {
	cmd := exec.Command("brew", "--version")
	if err := cmd.Run(); err == nil {
		return nil
	}

	cmd = exec.Command("/bin/bash", "-c", "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to install Homebrew: %v", err)
	}

	return nil
}
