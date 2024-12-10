//go:build windows
// +build windows

package packagemanager

import (
	"os/exec"
)

func Find(osid string) PackageManager {
	_, err := exec.LookPath("choco")
	if err == nil {
		return NewChocolatey(osid)
	}

	return nil
}
