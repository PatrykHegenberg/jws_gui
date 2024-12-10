//go:build darwin
// +build darwin

package packagemanager

import (
	"os/exec"
)

func Find(osid string) PackageManager {
	_, err := exec.LookPath("brew")
	if err == nil {
		return NewHomebrew(osid)
	}

	return nil
}
