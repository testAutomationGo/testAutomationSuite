package accountManagementUI

import (
	"testAutomationSuiteGO/internal/apiFunctions"
	"testAutomationSuiteGO/internal/testRunParameters"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type AccountManagementTab struct {
	name    string
	content fyne.CanvasObject
}

type CustomEntry struct {
	widget.Entry
	window fyne.Window
}

func NewCustomEntry(window fyne.Window) *CustomEntry {
	entry := &CustomEntry{window: window}
	entry.ExtendBaseWidget(entry)
	return entry
}

func DeleteUser(jwt string) (int, string) {
	code, response := apiFunctions.DeleteRequest(testRunParameters.GetApiEndpoint()+"v3/users", jwt)
	return code, response
}
