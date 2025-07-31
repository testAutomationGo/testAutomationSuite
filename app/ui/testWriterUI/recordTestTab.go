package testWriterUI

import (
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/ui/runParametersForUI"
	"testAutomationSuiteGO/app/uiFunctions"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func RecordTestTab() fyne.CanvasObject {

	recordTestLabel := widget.NewLabel("Record Test")

	enterTestNameLabel := widget.NewLabel("Enter Test Name")

	enterTestNameEntry := widget.NewEntry()

	testTypesLabel := widget.NewLabel("Select Test Type")
	testTypesOptions := runParametersForUI.GetTestSectorNames()
	selectTestType := widget.NewSelect(testTypesOptions, func(value string) {})
	selectTestType.SetSelectedIndex(0)

	recordTestButton := widget.NewButton("Record Test", func() {
		if enterTestNameEntry.Text == "" {
			uiFunctions.UpdateGUIConsole(TestWriterConsoleOutput, "Please enter a test name.")
			return
		}
		go RecordTest(shared.GetCurrentEnvironment(), enterTestNameEntry.Text, selectTestType.Selected)
	})

	clearConsoleBottomButton := widget.NewButton("Clear Console", func() {
		uiFunctions.UpdateGUIConsole(TestWriterConsoleOutput, "")
	})

	leftContent := container.NewVBox(
		recordTestLabel,
		enterTestNameLabel,
		enterTestNameEntry,
		testTypesLabel,
		selectTestType,
		recordTestButton,
		layout.NewSpacer(),
		clearConsoleBottomButton,
	)

	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, TestWriterConsoleOutput)

	return contentWithConsole
}
