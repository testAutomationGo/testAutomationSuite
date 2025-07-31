package queryDB

import (
	"testAutomationSuiteGO/app/shared"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

var UpdateConsoleOutput *CustomEntry

func UpdateTab(deps shared.AppDependencies) fyne.CanvasObject {

	UpdateConsoleOutput = NewCustomEntry(deps.MainWindow)
	UpdateConsoleOutput.SetPlaceHolder("Update Console Output")

	consolePadded := container.NewPadded(UpdateConsoleOutput)

	topControls := container.NewVBox()

	split := container.NewVSplit(container.NewPadded(topControls), consolePadded)

	contentWithConsole := container.NewBorder(nil, nil, nil, nil, split)

	return contentWithConsole
}
