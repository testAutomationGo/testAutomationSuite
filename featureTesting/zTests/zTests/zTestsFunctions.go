package zTests

import (
	"fmt"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/uiFunctions"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func GenerateWindowContent() fyne.CanvasObject {

	consoleOutput := uiFunctions.NewCustomEntry(shared.GetThisWindow())
	appendToConsole := func(text string) {
		currentText := consoleOutput.Text
		consoleOutput.SetText(fmt.Sprintf("%s\n%s", currentText, text))
	}
	leftContent := container.NewVBox(
		widget.NewButton("Run Tests", func() {
			appendToConsole("Running tests...")
			appendToConsole("Complete")

		}),
	)

	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, consoleOutput)

	return contentWithConsole
}
