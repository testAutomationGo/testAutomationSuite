package genStructFromDatabase

import (
	"os"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func GenStructFromDatabaseWindow(deps shared.AppDependencies) fyne.CanvasObject {
	pathSuffix := "/testData/codeRunnerMain/main"
	if !testingToolkit.VerifyFolderIsPresent(testingToolkit.CurrPath() + pathSuffix) {
		os.MkdirAll(testingToolkit.CurrPath()+pathSuffix, 0755)
	}
	consoleOutput := widget.NewMultiLineEntry()
	consoleOutput.SetPlaceHolder("Console output will appear here...")
	appendToConsole := func(text string) {
		fyne.Do(func() {
			consoleOutput.SetText(consoleOutput.Text + text + "\n")
		})
	}

	runGoCodeButton := widget.NewButton("Run Go Code", func() {
		go func() {
			appendToConsole("Running Go code...")
			code := consoleOutput.Text
			appendToConsole("Code:\n" + code)
		}()
	})

	bottomClearConsoleButton := widget.NewButton("Clear Console", func() {
		fyne.Do(func() {
			consoleOutput.SetText("")
		})
	})

	leftContent := container.NewVBox(
		runGoCodeButton,
		layout.NewSpacer(),
		bottomClearConsoleButton,
	)
	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, consoleOutput)
	return contentWithConsole
}
