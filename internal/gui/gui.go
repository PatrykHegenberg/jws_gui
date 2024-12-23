package gui

import (
	"log"

	"fyne.io/fyne/theme"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/PatrykHegenberg/jws_gui/internal/platform"
)

func createDependencyList(pm *platform.PlatformManager) *widget.List {
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
			req.InstalledBind.AddListener(binding.NewDataListener(func() {
				installed, _ := req.InstalledBind.Get()
				if installed {
					icon.SetResource(theme.ConfirmIcon())
				} else {
					icon.SetResource(theme.CancelIcon())
				}
			}))
		},
	)
	return list
}

func createProjectBox(pm *platform.PlatformManager) *fyne.Container {
	return container.NewVBox(
		widget.NewButton("Basic JakartaEE with Servlet and DB", func() {
			log.Println("Basic JakartaEE with Servlet")
			log.Println(pm.OS.Name)
		}),
		widget.NewButton("Basic JakartaEE with JSF and DB", func() {
			log.Println("Basic JakartaEE Project")
			log.Println(pm.OS.Name)
		}),
		widget.NewButton("Basic JakartaEE with RestAPI and DB", func() {
			log.Println("JakartaEE Project with RestAPI")
			log.Println(pm.OS.Name)
		}),
		widget.NewButton("Basic SpringBoot MicroService with DB", func() {
			log.Println("SpringBoot Project")
			log.Println(pm.OS.Name)
		}),
	)
}

func createTitle() *fyne.Container {
	titleLabel := widget.NewLabel("Java Web Dev Starter")
	titleLabel.TextStyle = fyne.TextStyle{
		Bold:   true,
		Italic: false,
	}
	titleLabel.Resize(fyne.NewSize(400, 50))
	return container.NewCenter(titleLabel)
}

func SetupGUI(pm *platform.PlatformManager) {
	myApp := app.New()
	myWindow := myApp.NewWindow("Uni Project Starter")

	titleContainer := createTitle()

	list := createDependencyList(pm)

	updateList := func() {
		list.Refresh()
	}

	installButton := widget.NewButton("Fehlende Pakete installieren", func() {
		err := pm.CheckAndInstallRequirements(true, myWindow)
		if err != nil {
			dialog.ShowError(err, myWindow)
		}
		list.Refresh()
		updateList()
	})

	packageBox := container.NewBorder(nil, installButton, nil, nil, list)
	projectsBox := createProjectBox(pm)

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

	updateList()
	pm.AllInstalled.AddListener(binding.NewDataListener(func() {
		allInstalled, _ := pm.AllInstalled.Get()
		if allInstalled {
			list.Hide()
			installButton.Hide()
			projectsBox.Show()
		} else {
			list.Show()
			installButton.Show()
			projectsBox.Hide()
		}
	}))
	myWindow.SetContent(content)

	myWindow.Resize(fyne.NewSize(800, 400))
	myWindow.ShowAndRun()
}
