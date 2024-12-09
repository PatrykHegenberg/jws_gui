package operatingsystem

import (
	"os/exec"
	"strings"
)

func getSysctlValue(key string) (string, error) {
	stdout, err := exec.Command(".", "sysctl", key).Output()
	if err != nil {
		return "", err
	}
	version := strings.TrimPrefix(string(stdout), key+": ")
	return strings.TrimSpace(version), nil
}

func platformInfo() (*OS, error) {
	var result OS
	result.ID = "Unknown"
	result.Name = "MacOS"
	result.Version = "Unknown"

	version, err := getSysctlValue("kern.osproductversion")
	if err != nil {
		return nil, err
	}
	result.Version = version
	ID, err := getSysctlValue("kern.osversion")
	if err != nil {
		return nil, err
	}
	result.ID = ID

	return &result, nil
}
