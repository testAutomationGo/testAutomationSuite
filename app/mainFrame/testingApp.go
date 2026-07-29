package mainFrame

import (
	"fmt"
	"log"
	"os"

	"testAutomationSuiteGO/app/rebuildApp"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/ui/featureTestsUI"
	"testAutomationSuiteGO/app/ui/fileCreatorUI"
	"testAutomationSuiteGO/app/ui/testsUI"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func AppMainMenu(myWindow fyne.Window, deps shared.AppDependencies) *fyne.MainMenu {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("New", func() {

		}),
		fyne.NewMenuItem("Open", func() {

		}),
		fyne.NewMenuItem("Save", func() {

		}),
	)

	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Cut", func() {

		}),
		fyne.NewMenuItem("Copy", func() {

		}),
		fyne.NewMenuItem("Paste", func() {

		}),
		fyne.NewMenuItem("Find", func() {
			ShowFindDialog(myWindow)
		}),
	)

	settingMenu := fyne.NewMenu("Settings",
		fyne.NewMenuItem("Preferences", func() {
			ShowPreferencesDialog(deps)
		}),
		fyne.NewMenuItem("Appearance", func() {

		}),
	)

	toolsMenu := fyne.NewMenu("Tools",
		fyne.NewMenuItem("Rebuild The App", func() {
			log.Println("Rebuilding the app.")
			buildString := rebuildApp.RebuildDesktopApp()
			log.Println("Build String: " + buildString)
			rebuildApp.ShutDownNotification(deps)
		}),
		fyne.NewMenuItem("Clear Output Folder", func() {
			folderPath := testingToolkit.CurrPath() + "/testData/appOutput"
			warningDialog := dialog.NewConfirm("Warning", "The folder will be completely emptied. Do you want to continue?", func(confirm bool) {
				if confirm {
					files, err := os.ReadDir(folderPath)
					if err != nil {
						log.Println("Could not read directory:", err)
						fmt.Println("failed to read directory: %w", err)
					}

					for _, file := range files {
						filePath := folderPath + "/" + file.Name()
						if file.IsDir() {
							testingToolkit.DeleteFolder(filePath)
						} else {
							if err := os.Remove(filePath); err != nil {
								log.Println("Could not remove file:", err)
								fmt.Println("failed to remove file: %w", err)
							}
						}
					}
				}
			}, myWindow)
			warningDialog.Show()
		}),
		fyne.NewMenuItem("Open Output Folder", func() {
			folderPath := testingToolkit.CurrPath() + "/testData/appOutput"
			if err := OpenFolder(folderPath); err != nil {
				log.Println("Could not open folder:", err)
				fmt.Println("failed to open folder: %w", err)
			}
		}),
		fyne.NewMenuItem("Updata QueryDB Schema", func() {
			uiFunctions.UpdateDBSchemas(myWindow, deps)
		}),
	)

	terminalMenu := fyne.NewMenu("Terminal",
		fyne.NewMenuItem("Open Terminal", func() {
		}),
		fyne.NewMenuItem("Run Command", func() {
		}),
	)

	mainMenu := fyne.NewMainMenu(
		fileMenu,
		editMenu,
		settingMenu,
		toolsMenu,
		terminalMenu,
	)
	return mainMenu
}

func AppWindow(myApp fyne.App) fyne.Window {
	myWindow := myApp.NewWindow("Testing App")
	return myWindow
}

func SetupTabs(deps shared.AppDependencies) {

	tabs := []Tab{
		{
			name:    "Run Test Sets",
			content: testsUI.RunFullTestSets(deps),
		}, /*
			{
				name:    "Test Cases",
				content: testCasesUI.TestCasesWindow(deps),
			},
			{
				name:    "Apps Navigations",
				content: appsNavigation.AppsNavigationWindow(deps),
			},
			{
				name:    "Api Requests",
				content: apiUI.ApiRequestsWindow(deps),
			},
			{
				name:    "Page Tools",
				content: pageToolsUI.PageToolsWindow(deps),
			},
			{
				name:    "Test Data Manager",
				content: testDataManagerUI.TestDataManagerWindow(deps),
			},
			{
				name:    "Note Taker",
				content: noteTakerUI.NoteTakerWindow(deps),
			},*/
		{
			name:    "File Creator",
			content: fileCreatorUI.FileCreatorWindow(deps),
		},
		/*{
			name:    "Query DB",
			content: queryDB.QueryDBWindow(deps),
		},*/
		{
			name:    "Feature Tests",
			content: featureTestsUI.FeatureTestsWindow(),
		},
	}

	tabItems := make([]*container.TabItem, len(tabs))
	for i, tab := range tabs {
		tabItems[i] = container.NewTabItem(tab.name, tab.content)
	}

	appTabs := container.NewAppTabs(tabItems...)

	topBar := topBar(deps)

	mainContent := container.NewBorder(topBar, nil, nil, nil, appTabs)

	deps.MainWindow.SetMainMenu(AppMainMenu(deps.MainWindow, deps))
	deps.MainWindow.SetContent(mainContent)

	deps.MainWindow.Resize(fyne.NewSize(1000, 700))
	deps.MainWindow.CenterOnScreen()
	deps.MainWindow.ShowAndRun()
}

func topBar(deps shared.AppDependencies) fyne.CanvasObject {
	envOptions := shared.EnvOptions
	envSelect := widget.NewSelect(envOptions, func(value string) {
	})
	envSelect.SetSelectedIndex(0)
	envSelect.OnChanged = func(value string) {
		envSelect.SelectedIndex()
		shared.SetCurrentEnvironment(shared.GetAppConfigFromFile().Envs[envSelect.SelectedIndex()])
	}

	envLabel := widget.NewLabel("Select Environment:")

	envContainer := container.NewHBox(
		envLabel,
		envSelect,
	)

	var SetJWTForAnyUserButton *widget.Button

	updateUserButtonText := func() {
		if shared.GetCurrentUserJWT() == "No JWT set" {
			SetJWTForAnyUserButton.SetText("Set JWT")
			return
		} else {
			SetJWTForAnyUserButton.SetText("Change JWT")
		}
	}

	userJWTLabel := widget.NewLabel("User JWT:")

	SetJWTForAnyUserButton = widget.NewButton("Set JWT", func() {
		jwtEntry := widget.NewEntry()
		d := dialog.NewCustom("Set Users JWT", "Cancel", container.NewVBox(
			widget.NewLabel("                  Enter JWT:                  "),
			jwtEntry,
		), deps.MainWindow)

		setButton := widget.NewButton("   Set   ", func() {
			shared.SetCurrentUserJWT(jwtEntry.Text)
			d.Hide()

			updateUserButtonText()
		})

		d.SetButtons([]fyne.CanvasObject{
			setButton,
		})

		d.Show()
	})
	SetJWTForAnyUserButton.Importance = widget.LowImportance

	useUserJWTCheck := widget.NewCheck("Use User JWT", nil)
	useUserJWTCheck.OnChanged = func(isChecked bool) {
		if isChecked {
			shared.SetUseUserJWT(true)
		} else {
			shared.SetUseUserJWT(false)
		}
	}

	userContainer := container.NewHBox(
		userJWTLabel,
		SetJWTForAnyUserButton,
		useUserJWTCheck,
	)

	return container.NewBorder(nil, nil, envContainer, userContainer)
}
