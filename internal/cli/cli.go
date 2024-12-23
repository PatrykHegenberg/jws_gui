package cli

import (
	"fmt"
	"log"

	"github.com/PatrykHegenberg/jws_gui/internal/platform"
	"github.com/spf13/cobra"
)

func SetupCLI(pm *platform.PlatformManager) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "uni-project-starter",
		Short: "Universitäts-Projekt-Starter-Anwendung",
	}

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Überprüft Systemanforderungen",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Erkannter Paketmanager: %s\n", pm.PackageManager.Name())

			for _, req := range pm.Requirements {
				status := "nicht installiert"
				if req.Installed {
					status = "installiert"
				}
				fmt.Printf("%s: %s\n", req.Name, status)
			}
		},
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Installiert fehlende Systemanforderungen",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pm.CheckAndInstallRequirements(false, nil); err != nil {
				log.Fatalf("Fehler bei der Installation: %v", err)
			}
		},
	}

	rootCmd.AddCommand(checkCmd, installCmd)
	return rootCmd
}
