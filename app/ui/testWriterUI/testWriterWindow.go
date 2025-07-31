package testWriterUI

import (
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/uiFunctions"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var TestWriterConsoleOutput *widget.Entry

func TestWriterWindow(deps shared.AppDependencies) fyne.CanvasObject {

	consoleOutput := widget.NewMultiLineEntry()
	consoleOutput.SetPlaceHolder(TestWriterConsoleRecordTestInstructions())
	TestWriterConsoleOutput = consoleOutput

	testDB = CheckOnAndCreateDBAndTable()
	if testDB == nil {

		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not create database.")
	}

	tabs := []TestWriterTab{
		{
			name:    "Record Test",
			content: RecordTestTab(),
		},
		{
			name:    "Replay And Edit",
			content: ReplayTestTab(),
		},
	}

	tabItems := make([]*container.TabItem, len(tabs))

	for i, tab := range tabs {
		tabItems[i] = container.NewTabItem(tab.name, tab.content)
	}

	tabsContainer := container.NewAppTabs(tabItems...)

	return container.NewBorder(nil, nil, nil, nil, tabsContainer)
}
