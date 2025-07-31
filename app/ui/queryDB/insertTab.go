package queryDB

import (
	"testAutomationSuiteGO/app/shared"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

var InsertConsoleOutput *CustomEntry

func InsertTab(deps shared.AppDependencies) fyne.CanvasObject {

	InsertConsoleOutput = NewCustomEntry(deps.MainWindow)
	InsertConsoleOutput.SetPlaceHolder("Insert Console Output")

	consolePadded := container.NewPadded(InsertConsoleOutput)

	topControls := container.NewVBox()

	split := container.NewVSplit(container.NewPadded(topControls), consolePadded)

	contentWithConsole := container.NewBorder(nil, nil, nil, nil, split)

	return contentWithConsole
}
