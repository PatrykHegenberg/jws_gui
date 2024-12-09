package main

import (
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/PatrykHegenberg/jws_gui/internal/system/operatingsystem"
	"github.com/PatrykHegenberg/jws_gui/internal/system/packagemanager"
	"github.com/spf13/cobra"
)

var requiredPackages = map[string][]*packagemanager.Package{
	"git": {
		{
			Name:          "git",
			SystemPackage: true,
			InstallCommand: map[string]string{
				"apt":      "git",
				"dnf":      "git",
				"pacman":   "git",
				"zypper":   "git",
				"homebrew": "git",
				"choco":    "git",
			},
		},
	},
	"openjdk": {
		{
			Name:          "openjdk",
			SystemPackage: true,
			InstallCommand: map[string]string{
				"apt":      "openjdk-17-jdk",
				"dnf":      "java-17-openjdk-devel",
				"pacman":   "jdk17-openjdk",
				"zypper":   "java-17-openjdk-devel",
				"homebrew": "openjdk@17",
				"choco":    "openjdk",
			},
		},
	},
	"podman": {
		{
			Name:          "podman",
			SystemPackage: true,
			InstallCommand: map[string]string{
				"apt":      "podman",
				"dnf":      "podman",
				"pacman":   "podman",
				"zypper":   "podman",
				"homebrew": "podman",
				"choco":    "podman",
			},
		},
	},
	"vscode": {
		{
			Name:          "vscode",
			SystemPackage: true,
			InstallCommand: map[string]string{
				"apt":      "code",
				"dnf":      "code",
				"pacman":   "code",
				"zypper":   "code",
				"homebrew": "visual-studio-code",
				"choco":    "vscode",
			},
		},
	},
}

type SoftwareRequirement struct {
	Name           string
	Package        *packagemanager.Package
	InstallCommand string
	Installed      bool
}

type PlatformManager struct {
	PackageManager packagemanager.PackageManager
	Requirements   []*SoftwareRequirement
	OS             *operatingsystem.OS
}

func NewPlatformManager() *PlatformManager {
	pm := &PlatformManager{}

	osInfo, err := operatingsystem.Info()
	if err != nil {
		log.Fatalf("Konnte Betriebssysteminformationen nicht abrufen: %v", err)
	}
	pm.OS = osInfo

	pm.PackageManager = packagemanager.Find(osInfo.ID)
	if pm.PackageManager == nil {
		log.Fatal("Kein unterstützter Paketmanager gefunden")
	}

	pm.initRequirements()

	return pm
}

func (pm *PlatformManager) initRequirements() {
	for name, packages := range requiredPackages {
		for _, pkg := range packages {
			requirement := &SoftwareRequirement{
				Name:    name,
				Package: pkg,
			}

			available, err := pm.PackageManager.PackageAvailable(pkg)
			if err != nil || !available {
				log.Printf("Paket %s nicht verfügbar: %v", name, err)
				continue
			}

			installed, err := pm.PackageManager.PackageInstalled(pkg)
			if err != nil {
				log.Printf("Fehler bei Installationsprüfung für %s: %v", name, err)
				continue
			}

			requirement.Installed = installed
			requirement.InstallCommand = pm.PackageManager.InstallCommand(pkg)

			pm.Requirements = append(pm.Requirements, requirement)
		}
	}
}

func (pm *PlatformManager) checkAndInstallRequirements() error {
	for _, req := range pm.Requirements {
		if !req.Installed {
			fmt.Printf("Installiere %s...\n", req.Name)
			// Hier könnten Sie die tatsächliche Installationslogik implementieren
			// z.B. exec.Command(req.InstallCommand).Run()
		}
	}
	return nil
}

func setupCLI(pm *PlatformManager) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "uni-project-starter",
		Short: "Universitäts-Projekt-Starter-Anwendung",
	}

	var checkCmd = &cobra.Command{
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

	var installCmd = &cobra.Command{
		Use:   "install",
		Short: "Installiert fehlende Systemanforderungen",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pm.checkAndInstallRequirements(); err != nil {
				log.Fatalf("Fehler bei der Installation: %v", err)
			}
		},
	}

	rootCmd.AddCommand(checkCmd, installCmd)
	return rootCmd
}

func setupGUI(pm *PlatformManager) {
	myApp := app.New()
	myWindow := myApp.NewWindow("Uni Project Starter")

	content := container.NewVBox(
		widget.NewLabel("Universitäts Projekt Starter"),
		widget.NewButton("Systemcheck", func() {
			fmt.Printf("Paketmanager: %s\n", pm.PackageManager.Name())
			for _, req := range pm.Requirements {
				status := "nicht installiert"
				if req.Installed {
					status = "installiert"
				}
				fmt.Printf("%s: %s\n", req.Name, status)
			}
		}),
		widget.NewButton("Projekt erstellen", func() {
			// Projektgenerierungs-Logik
		}),
	)

	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(400, 300))
	myWindow.ShowAndRun()
}

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
