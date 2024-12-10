package platform

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/PatrykHegenberg/jws_gui/internal/system/operatingsystem"
	"github.com/PatrykHegenberg/jws_gui/internal/system/packagemanager"
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
	InstalledBind  binding.Bool
}

type PlatformManager struct {
	PackageManager packagemanager.PackageManager
	Requirements   []*SoftwareRequirement
	OS             *operatingsystem.OS
	AllInstalled   binding.Bool
}

func NewPlatformManager() *PlatformManager {
	pm := &PlatformManager{
		AllInstalled: binding.NewBool(),
	}

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
			requirement.InstalledBind = binding.NewBool()
			requirement.InstalledBind.Set(requirement.Installed)

			pm.Requirements = append(pm.Requirements, requirement)
		}
	}
}

func (pm *PlatformManager) CheckAndInstallRequirements(gui bool, window fyne.Window) error {
	if !gui {
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

	var queue []*SoftwareRequirement
	for _, req := range pm.Requirements {
		if !req.Installed {
			queue = append(queue, req)
		}
	}

	var processNext func()
	processNext = func() {
		if len(queue) == 0 {
			dialog.ShowInformation("Fertig", "Alle Pakete wurden verarbeitet.", window)
			return
		}

		req := queue[0]
		queue = queue[1:]

		dialog.ShowConfirm("Installation erforderlich",
			fmt.Sprintf("Möchten Sie %s installieren?", req.Name),
			func(install bool) {
				if install {
					passwordEntry := widget.NewPasswordEntry()
					dialog.ShowForm("Sudo-Passwort erforderlich", "OK", "Abbrechen",
						[]*widget.FormItem{widget.NewFormItem("Sudo Passwort", passwordEntry)},
						func(submitted bool) {
							if !submitted {
								dialog.ShowError(fmt.Errorf("Installation abgebrochen"), window)
								processNext()
								return
							}

							sudoPass := passwordEntry.Text
							if sudoPass == "" {
								dialog.ShowError(fmt.Errorf("Kein Passwort eingegeben"), window)
								processNext()
								return
							}

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
												req.InstalledBind.Set(true)
												pm.checkAllInstalled()
												return "erfolgreich installiert"
											}
											return fmt.Sprintf("mit Fehlern installiert: %v", err)
										}()),
								})
								processNext()
							}()
						}, window)
				} else {
					processNext()
				}
			}, window)
	}

	processNext()
	return nil
}

func (pm *PlatformManager) installPackage(pkg *packagemanager.Package, gui bool, window fyne.Window) error {
	if gui {
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
					pm.checkAllInstalled()
				}
			}, window)
		return nil
	}

	fmt.Print("Bitte geben Sie Ihr sudo-Passwort ein: ")
	passBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("Fehler beim Lesen des Passworts: %v", err)
	}
	sudoPass := string(passBytes)
	fmt.Println()

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

func (pm *PlatformManager) checkAllInstalled() {
	allInstalled := true
	for _, req := range pm.Requirements {
		if !req.Installed {
			allInstalled = false
			break
		}
	}
	pm.AllInstalled.Set(allInstalled)
}
