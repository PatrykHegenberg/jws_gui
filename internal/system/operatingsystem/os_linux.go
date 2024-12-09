//go:build linux
// +build linux

package operatingsystem

import (
	"fmt"
	"os"
	"strings"
)

func platformInfo() (*OS, error) {
	_, err := os.Stat("/etc/os-release")
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("unable to read system information")
	}

	osRelease, _ := os.ReadFile("/etc/os-release")
	return parseOsRelease(string(osRelease)), nil
}

func parseOsRelease(osRelease string) *OS {
	var result OS
	result.ID = "Unknown"
	result.Name = "Unknown"
	result.Version = "Unknown"

	lines := strings.Split(osRelease, "\n")

	for _, line := range lines {
		splitLine := strings.SplitN(line, "=", 2)
		if len(splitLine) != 2 {
			continue
		}
		switch splitLine[0] {
		case "ID":
			result.ID = strings.ToLower(strings.Trim(splitLine[1], "\""))
		case "NAME":
			result.Name = strings.Trim(splitLine[1], "\"")
		case "VERSION_ID":
			result.Version = strings.Trim(splitLine[1], "\"")
		}
	}
	return &result
}
