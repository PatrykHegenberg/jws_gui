//go:build windows
// +build windows

package packagemanager

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Chocolatey struct {
	name string
	osid string
}

func NewChocolatey(osid string) *Chocolatey {
	return &Chocolatey{
		name: "choco",
		osid: osid,
	}
}

func (c *Chocolatey) Packages() packagemap {
	universalPackages := GenerateUniversalPackages()
	
	// Add Windows-specific package modifications if needed
	windowsSpecificPackages := packagemap{
		"openjdk": {
			{
				Name:          "openjdk",
				SystemPackage: true,
				NativePackageName: map[string]string{
					"choco": "openjdk17",
				},
			},
		},
		"vscode": {
			{
				Name:          "vscode",
				SystemPackage: false,
				NativePackageName: map[string]string{
					"choco": "vscode",
				},
			},
		},
	}

	// Merge Windows-specific packages with universal packages
	for key, value := range windowsSpecificPackages {
		universalPackages[key] = value
	}

	return universalPackages
}

func (c *Chocolatey) InstallCommand(pkg *Package) string {
	// Use the package name specific to Chocolatey
	packageName := pkg.NativePackageName[c.name]
	
	// If no Chocolatey-specific name is found, fallback to the default
	if packageName == "" {
		packageName = pkg.NativePackageName[c.osid]
	}

	return fmt.Sprintf("choco install %s -y", packageName)
}

func (c *Chocolatey) Name() string {
	return c.name
}

func (c *Chocolatey) PackageInstalled(pkg *Package) (bool, error) {
	// Skip non-system packages
	if !pkg.SystemPackage {
		return false, nil
	}

	// Get the package name for Chocolatey
	packageName := pkg.NativePackageName[c.name]
	if packageName == "" {
		return false, fmt.Errorf("no Chocolatey package name found for %s", pkg.Name)
	}

	// Run choco list to check if package is installed
	cmd := exec.Command("choco", "list", "--local-only", packageName)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// Check if the output contains the package name
	return strings.Contains(string(output), packageName), nil
}

func (c *Chocolatey) PackageAvailable(pkg *Package) (bool, error) {
	// Skip non-system packages
	if !pkg.SystemPackage {
		return false, nil
	}

	// Get the package name for Chocolatey
	packageName := pkg.NativePackageName[c.name]
	if packageName == "" {
		return false, fmt.Errorf("no Chocolatey package name found for %s", pkg.Name)
	}

	// Run choco search to check package availability
	cmd := exec.Command("choco", "search", packageName)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// Use regex to find package and extract version
	reg := regexp.MustCompile(packageName + `\s+(\d+[\.\d+]*)`)
	matches := reg.FindStringSubmatch(string(output))
	
	if len(matches) > 1 {
		pkg.Version = matches[1]
		return true, nil
	}

	return false, nil
}

// EnsureInstalled checks if Chocolatey is installed, and if not, attempts to install it
func (c *Chocolatey) EnsureInstalled() error {
	// Check if Chocolatey is already installed
	_, err := exec.LookPath("choco")
	if err == nil {
		return nil // Chocolatey is already installed
	}

	// Attempt to install Chocolatey using PowerShell
	cmd := exec.Command("powershell", 
		"-NoProfile", 
		"-ExecutionPolicy", "Bypass", 
		"-Command",
		"Set-ExecutionPolicy Bypass -Scope Process -Force; " +
		"[System.Net.ServicePointManager]::SecurityProtocol = " +
		"[System.Net.ServicePointManager]::SecurityProtocol -bor 3072; " +
		"iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))")
	
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to install Chocolatey: %v", err)
	}

	return nil
}
