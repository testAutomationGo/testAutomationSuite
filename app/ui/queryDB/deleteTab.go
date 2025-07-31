package queryDB

import (
	"testAutomationSuiteGO/app/shared"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

var DeleteConsoleOutput *CustomEntry

func DeleteTab(deps shared.AppDependencies) fyne.CanvasObject {

	DeleteConsoleOutput = NewCustomEntry(deps.MainWindow)
	DeleteConsoleOutput.SetPlaceHolder("Delete Console Output")

	consolePadded := container.NewPadded(DeleteConsoleOutput)

	topControls := container.NewVBox()

	split := container.NewVSplit(container.NewPadded(topControls), consolePadded)

	contentWithConsole := container.NewBorder(nil, nil, nil, nil, split)

	return contentWithConsole
}
