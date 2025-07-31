package loggingUI

import (
	"log"
	"testAutomationSuiteGO/app/ui/runParametersForUI"
	"testAutomationSuiteGO/app/uiFunctions"

	"fyne.io/fyne/v2/widget"
)

func GuiLog(consoleOutput *widget.Entry, logStatement string, tcNumber string) {
	guiRun := runParametersForUI.GetGuiRun()
	if guiRun {
		uiFunctions.AppendToGUIConsole(consoleOutput, tcNumber+": "+logStatement)
		log.Println(tcNumber + ": " + logStatement)
	}
}
