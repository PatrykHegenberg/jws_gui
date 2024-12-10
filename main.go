package main

import (
	"log"
	"os"

	"github.com/PatrykHegenberg/jws_gui/internal/cli"
	"github.com/PatrykHegenberg/jws_gui/internal/gui"
	"github.com/PatrykHegenberg/jws_gui/internal/platform"
)

func main() {
	pm := platform.NewPlatformManager()

	if len(os.Args) > 1 {
		if err := cli.SetupCLI(pm).Execute(); err != nil {
			log.Fatal(err)
		}
	} else {
		gui.SetupGUI(pm)
	}
}
