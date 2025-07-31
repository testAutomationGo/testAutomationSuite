package queryDB

import (
	"testAutomationSuiteGO/app/shared"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

var QueryWriterOutputConsole *CustomEntry

func QueryWriterTab(deps shared.AppDependencies) fyne.CanvasObject {

	QueryWriterOutputConsole = NewCustomEntry(deps.MainWindow)
	QueryWriterOutputConsole.SetPlaceHolder("Console Output")

	consolePadded := container.NewPadded(QueryWriterOutputConsole)

	topControls := container.NewVBox()

	split := container.NewVSplit(container.NewPadded(topControls), consolePadded)

	contentWithConsole := container.NewBorder(nil, nil, nil, nil, split)

	return contentWithConsole
}
