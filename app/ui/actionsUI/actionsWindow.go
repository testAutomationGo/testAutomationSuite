package actionsUI

import (
	"fmt"
	"testAutomationSuiteGO/app/shared"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func ActionsWindow(deps shared.AppDependencies) fyne.CanvasObject {
	consoleOutput := widget.NewMultiLineEntry()
	consoleOutput.SetPlaceHolder("Console Output:")

	appendToConsole := func(text string) {
		currentText := consoleOutput.Text
		consoleOutput.SetText(fmt.Sprintf("%s\n%s", currentText, text))
	}

	testSectorOptions := GetTestSectorsOptions()

	var testSector string

	testSectorSelect := widget.NewSelect(testSectorOptions, func(value string) {
		testSector = value
	})
	testSectorSelect.SetSelectedIndex(0)
	testSectorBox := container.NewVBox(
		widget.NewLabel("Select Test Sector:"),
		testSectorSelect,
	)

	runV3ActionsButton := widget.NewButton("Run V3 Actions", func() {
		var selectedEnv string
		appendToConsole("Running V3 Actions...")
		if shared.GetCurrentEnvironment() == "" {
			appendToConsole("Please select an environment.")
			return
		}
		if testSector == "" {
			appendToConsole("Please select a test sector.")
			return
		}
		switch shared.GetCurrentEnvironment() {
		case "DEV":
			selectedEnv = "develop"
		case "PRD":
			selectedEnv = "production"
		default:
			selectedEnv = "develop"
		}
		inputs := map[string]string{
			"environment": selectedEnv,
			"test_area":   testSector,
		}
		appendToConsole(fmt.Sprintf("Selected Environment: %s", selectedEnv))
		appendToConsole(fmt.Sprintf("Selected Test Sector: %s", testSector))
		code, response := RunActions(GetGitHubActionsToken(), "PinataCloud", "testAutomationSuiteGO", "run-tests-V3.yml", inputs)
		appendToConsole(fmt.Sprintf("Response Code: %d", code))
		if code == 204 {
			appendToConsole("Workflow triggered successfully!")
		} else {
			appendToConsole("Workflow trigger failed.")
			appendToConsole(fmt.Sprintf("Error: received status code %d, response: %s", code, response))
		}

	})

	bottomClearConsoleButton := widget.NewButton("     Clear Console     ", func() {
		go consoleOutput.SetText("")
	})

	leftContent := container.NewVBox(
		testSectorBox,
		runV3ActionsButton,
		layout.NewSpacer(),
		bottomClearConsoleButton,
	)

	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, consoleOutput)

	return contentWithConsole
}
