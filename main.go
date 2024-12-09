package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"fyne.io/fyne/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/PatrykHegenberg/jws_gui/internal/system/operatingsystem"
	"github.com/PatrykHegenberg/jws_gui/internal/system/packagemanager"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var requiredPackages = map[string][]*packagemanager.Package{
	"git": {
		{
			Name:          "git",
			SystemPackage: true,
			NativePackageName: map[string]string{
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
			NativePackageName: map[string]string{
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
			NativePackageName: map[string]string{
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
			NativePackageName: map[string]string{
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

func (pm *PlatformManager) checkAndInstallRequirements(gui bool, window fyne.Window) error {
	if !gui {
		// CLI-Modus: Synchron arbeiten
		for _, req := range pm.Requirements {
			if !req.Installed {
				fmt.Printf("Möchten Sie %s installieren? (j/n): ", req.Name)
				var response string
				fmt.Scanln(&response)

				if strings.ToLower(response) == "j" {
					err := pm.installPackage(req.Package, false, nil)
					if err != nil {
						return fmt.Errorf("Fehler bei Installation von %s: %v", req.Name, err)
					}
					req.Installed = true
				}
			}
		}
		return nil
	}

	// GUI-Modus: Asynchrone, sequentielle Verarbeitung
	var queue []*SoftwareRequirement
	for _, req := range pm.Requirements {
		if !req.Installed {
			queue = append(queue, req)
		}
	}

	var processNext func()
	processNext = func() {
		if len(queue) == 0 {
			// Alle Pakete abgearbeitet
			dialog.ShowInformation("Fertig", "Alle Pakete wurden verarbeitet.", window)
			return
		}

		// Nächstes Paket aus der Warteschlange holen
		req := queue[0]
		queue = queue[1:]

		dialog.ShowConfirm("Installation erforderlich",
			fmt.Sprintf("Möchten Sie %s installieren?", req.Name),
			func(install bool) {
				if install {
					// Passwortabfrage-Dialog starten
					passwordEntry := widget.NewPasswordEntry()
					dialog.ShowForm("Sudo-Passwort erforderlich", "OK", "Abbrechen",
						[]*widget.FormItem{widget.NewFormItem("Sudo Passwort", passwordEntry)},
						func(submitted bool) {
							if !submitted {
								dialog.ShowError(fmt.Errorf("Installation abgebrochen"), window)
								processNext() // Nächstes Paket abarbeiten
								return
							}

							sudoPass := passwordEntry.Text
							if sudoPass == "" {
								dialog.ShowError(fmt.Errorf("Kein Passwort eingegeben"), window)
								processNext() // Nächstes Paket abarbeiten
								return
							}

							// Installationskommando ausführen
							go func() {
								installCommand := pm.PackageManager.InstallCommand(req.Package)
								cmd := exec.Command("sudo", "-S", "sh", "-c", installCommand)
								cmd.Stdin = strings.NewReader(sudoPass + "\n")
								cmd.Stdout = os.Stdout
								cmd.Stderr = os.Stderr

								err := cmd.Run()
								fyne.CurrentApp().SendNotification(&fyne.Notification{
									Title: "Installation abgeschlossen",
									Content: fmt.Sprintf("Paket %s wurde %s",
										req.Name,
										func() string {
											if err == nil {
												req.Installed = true
												return "erfolgreich installiert"
											}
											return fmt.Sprintf("mit Fehlern installiert: %v", err)
										}()),
								})
								processNext() // Nächstes Paket abarbeiten
							}()
						}, window)
				} else {
					processNext() // Nächstes Paket abarbeiten
				}
			}, window)
	}

	processNext() // Starte die Verarbeitung
	return nil
}

func (pm *PlatformManager) installPackage(pkg *packagemanager.Package, gui bool, window fyne.Window) error {
	if gui {
		// GUI-Modus: Verwenden Sie einen Fyne-Dialog
		passwordEntry := widget.NewPasswordEntry()
		dialog.ShowForm("Sudo-Passwort erforderlich", "OK", "Abbrechen",
			[]*widget.FormItem{widget.NewFormItem("Sudo Passwort", passwordEntry)},
			func(submitted bool) {
				if !submitted {
					dialog.ShowError(fmt.Errorf("Installation abgebrochen"), window)
					return
				}

				sudoPass := passwordEntry.Text
				if sudoPass == "" {
					dialog.ShowError(fmt.Errorf("Kein Passwort eingegeben"), window)
					return
				}

				// Führe die Installation mit sudo aus
				installCommand := pm.PackageManager.InstallCommand(pkg)
				cmd := exec.Command("sudo", "-S", "sh", "-c", installCommand)
				cmd.Stdin = strings.NewReader(sudoPass + "\n")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				err := cmd.Run()
				if err != nil {
					dialog.ShowError(fmt.Errorf("Fehler bei Installation von %s: %v", pkg.Name, err), window)
				} else {
					dialog.ShowInformation("Installation erfolgreich",
						fmt.Sprintf("%s wurde erfolgreich installiert", pkg.Name), window)
				}
			}, window)
		return nil
	}

	// CLI-Modus: Fragen Sie nach dem Passwort in der Konsole
	fmt.Print("Bitte geben Sie Ihr sudo-Passwort ein: ")
	passBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("Fehler beim Lesen des Passworts: %v", err)
	}
	sudoPass := string(passBytes)
	fmt.Println() // Neue Zeile nach der Passworteingabe

	if sudoPass == "" {
		return fmt.Errorf("Installation abgebrochen")
	}

	installCommand := pm.PackageManager.InstallCommand(pkg)
	cmd := exec.Command("sudo", "-S", "sh", "-c", installCommand)
	cmd.Stdin = strings.NewReader(sudoPass + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Fehler bei Installation von %s: %v", pkg.Name, err)
	}

	return nil
}

func (pm *PlatformManager) checkAllInstalled() bool {
	for _, req := range pm.Requirements {
		if !req.Installed {
			return false
		}
	}
	return true
}

func setupCLI(pm *PlatformManager) *cobra.Command {
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
			if err := pm.checkAndInstallRequirements(false, nil); err != nil {
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

	titleLabel := widget.NewLabel("Java Web Dev Starter")

	titleLabel.TextStyle = fyne.TextStyle{
		Bold:   true,
		Italic: false,
	}

	titleLabel.Resize(fyne.NewSize(400, 50))

	titleContainer := container.NewCenter(titleLabel)

	list := widget.NewList(
		func() int { return len(pm.Requirements) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.ConfirmIcon()),
				widget.NewLabel("Template"),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			req := pm.Requirements[id]
			box := item.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			label := box.Objects[1].(*widget.Label)

			label.SetText(req.Name)
			if req.Installed {
				icon.SetResource(theme.ConfirmIcon())
			} else {
				icon.SetResource(theme.CancelIcon())
			}
		},
	)

	updateList := func() {
		list.Refresh()
	}

	installButton := widget.NewButton("Fehlende Pakete installieren", func() {
		err := pm.checkAndInstallRequirements(true, myWindow)
		if err != nil {
			dialog.ShowError(err, myWindow)
		}
		list.Refresh()
		updateList()
	})

	checkAllInstalled := func() bool {
		for _, req := range pm.Requirements {
			if !req.Installed {
				return false
			}
		}
		return true
	}

	packageBox := container.NewBorder(nil, installButton, nil, nil, list)

	projectsBox := container.NewVBox(
		widget.NewButton("Basic JakartaEE with Servlet and DB", func() {}),
		widget.NewButton("Basic JakartaEE with JSF and DB", func() {}),
		widget.NewButton("Basic JakartaEE with RestAPI and DB", func() {}),
		widget.NewButton("Basich SpringBoot MicroService with DB", func() {}),
	)
	content := container.NewBorder(
		titleContainer,
		projectsBox,
		nil,
		nil,
		container.NewBorder(
			nil,
			nil,
			nil,
			nil,
			packageBox,
		),
	)
	myWindow.SetContent(content)

	updateList()

	if checkAllInstalled() {
		installButton.Hide()
		projectsBox.Show()
	} else {
		installButton.Show()
		projectsBox.Hide()
	}

	myWindow.Resize(fyne.NewSize(800, 400))
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
