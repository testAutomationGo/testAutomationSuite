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

func AccountsTab(deps shared.AppDependencies) fyne.CanvasObject {
	// Create a new entry widget
	accountsTabConsole := NewCustomEntry(deps.MainWindow)

	// Function to append text to the console
	appendToConsole := func(text string) {
		currentText := accountsTabConsole.Text
		accountsTabConsole.SetText(fmt.Sprintf("%s\n%s", currentText, text))
	}

	var jwtForAccounts string

	SetJWTForAccountsButton := widget.NewButton("Set JWT", func() {
		jwtEntry := widget.NewEntry()
		d := dialog.NewCustom("Set Users JWT", "Cancel", container.NewVBox(
			widget.NewLabel("                  Enter JWT:                  "),
			jwtEntry,
		), deps.MainWindow)
		setButton := widget.NewButton("            Set            ", func() {
			jwtForAccounts = jwtEntry.Text
			if jwtForAccounts == "" {
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

	UserToDeleteLabel := widget.NewLabel("Set Clerk Token To Delete User:")
	UserToDeleteButton := widget.NewButton("Delete This User", func() {

		selectedEnvInt := uiFunctions.GetSelectIndex(shared.EnvOptions, shared.GetCurrentEnvironment())
		runParametersForUI.SetGUIEnvConfigurationsForAppFunctions(runParametersForUI.GetEnvironments()[selectedEnvInt])
		appendToConsole("User delete in progress in " + shared.GetCurrentEnvironment() + ".")
		deleteCode, deleteResponse := DeleteUser(jwtForAccounts)
		if deleteCode == 200 {
			appendToConsole("User deleted successfully.")
			appendToConsole("Response: " + deleteResponse)
		} else {
			appendToConsole("Failed to delete user.")
			appendToConsole("Response: " + deleteResponse)
		}

	})

	labelSpacer := widget.NewLabel("            ")

	bottomClearConsoleButton := widget.NewButton("Clear Console", func() {
		accountsTabConsole.SetText("")
	})

	leftContent := container.NewVBox(
		SetJWTForAccountsButton,
		labelSpacer,
		UserToDeleteLabel,
		UserToDeleteButton,
		layout.NewSpacer(),
		bottomClearConsoleButton,
	)

	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, accountsTabConsole)

	return contentWithConsole

}

// 	})
