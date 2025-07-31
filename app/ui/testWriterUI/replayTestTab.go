package testWriterUI

import (
	"testAutomationSuiteGO/app/uiFunctions"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var rightContent *fyne.Container

func ReplayTestTab() fyne.CanvasObject {

	replayTestLabel := widget.NewLabel("Replay Test")

	envSelectLabel := widget.NewLabel("Select Environment")

	testNames, err := GetTestNames(testDB)

	if err != nil {
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not get test names: "+err.Error())
		options := []string{"No tests"}
		return widget.NewSelect(options, func(selected string) {})
	}
	WrittenTestsSelect = widget.NewSelect(testNames, func(selected string) {})
	WrittenTestsSelect.PlaceHolder = "Select a test"

	enterTestNameLabel := widget.NewLabel("Select Test")

	replayTestButton := widget.NewButton("Replay Test", func() {
		go ReplayTest("testWriterEnvSelect.Selected", WrittenTestsSelect.Selected)
	})

	removeTestButton := widget.NewButton("Remove Test", func() {
		go RemoveTest(WrittenTestsSelect.Selected)
	})

	editTestButton := widget.NewButton("Edit Test", func() {
		go EditTestView(WrittenTestsSelect.Selected)
	})

	clearConsoleBottomButton := widget.NewButton("Clear Console", func() {
		uiFunctions.UpdateGUIConsole(TestWriterConsoleOutput, "")
	})

	leftContent := container.NewVBox(
		replayTestLabel,
		envSelectLabel,
		enterTestNameLabel,
		WrittenTestsSelect,
		replayTestButton,
		removeTestButton,
		editTestButton,
		layout.NewSpacer(),
		clearConsoleBottomButton,
	)

	rightContent = container.NewVBox()

	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, rightContent)

	return contentWithConsole
}
