package main

import (
	"log"
	"os"
)

func main() {
	pm := NewPlatformManager()

	if len(os.Args) > 1 {
		if err := setupCLI(pm).Execute(); err != nil {
			log.Fatal(err)
		}
	} else {
		setupGUI(pm)
	}
}
