package accountManagementUI

import (
	"fmt"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/ui/runParametersForUI"
	"testAutomationSuiteGO/app/uiFunctions"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func WorkspacesTab(deps shared.AppDependencies) fyne.CanvasObject {

	userSubTabConsole := NewCustomEntry(deps.MainWindow)

	appendToConsole := func(text string) {
		currentText := userSubTabConsole.Text
		userSubTabConsole.SetText(fmt.Sprintf("%s\n%s", currentText, text))
	}

	var jwtForWorkspaces string
	SetJWTForWorkspaceButton := widget.NewButton("Set JWT", func() {
		jwtEntry := widget.NewEntry()
		d := dialog.NewCustom("Set Users JWT", "Cancel", container.NewVBox(
			widget.NewLabel("                  Enter JWT:                  "),
			jwtEntry,
		), deps.MainWindow)
		setButton := widget.NewButton("            Set            ", func() {
			jwtForWorkspaces = jwtEntry.Text
			if jwtForWorkspaces == "" {
				appendToConsole("JWT is empty.")
			} else {
				appendToConsole("JWT set successfully.")
			}
			d.Hide()
		})
		d.SetButtons([]fyne.CanvasObject{
			setButton,
		})

		d.Show()
	})

	UserIDToDeleteLabel := widget.NewLabel("User ID to delete:")
	UserIDToDeleteEntry := widget.NewEntry()
	UserIDToDeleteEntry.SetPlaceHolder("Enter User ID")

	UserToDeleteButton := widget.NewButton("Delete User", func() {
		userID := UserIDToDeleteEntry.Text
		if userID == "" {
			appendToConsole("User ID is empty.")
		} else {
			selectedEnvInt := uiFunctions.GetSelectIndex(shared.EnvOptions, shared.GetCurrentEnvironment())
			runParametersForUI.SetGUIEnvConfigurationsForAppFunctions(runParametersForUI.GetEnvironments()[selectedEnvInt])
			appendToConsole("User ID to delete in " + shared.GetCurrentEnvironment() + ": " + userID)
			deleteCode, deleteResponse := DeleteUser(jwtForWorkspaces)
			if deleteCode == 200 {
				appendToConsole("User deleted successfully: " + userID)
				appendToConsole("Response: " + deleteResponse)
			} else {
				appendToConsole("Failed to delete user: " + userID)
				appendToConsole("Response: " + deleteResponse)
			}
		}
	})

	labelSpacer := widget.NewLabel("            ")

	bottomClearConsoleButton := widget.NewButton("Clear Console", func() {
		userSubTabConsole.SetText("")
	})

	leftContent := container.NewVBox(
		SetJWTForWorkspaceButton,
		labelSpacer,
		UserIDToDeleteLabel,
		UserIDToDeleteEntry,
		UserToDeleteButton,
		layout.NewSpacer(),
		bottomClearConsoleButton,
	)

	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, userSubTabConsole)

	return contentWithConsole

}
