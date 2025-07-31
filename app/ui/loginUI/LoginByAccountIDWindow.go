package loginUI

import (
	"fmt"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/ui/runParametersForUI"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var (
	LoginUIOutputConsole *widget.Entry
)

func LoginByAccountIDWindow(deps shared.AppDependencies) fyne.CanvasObject {

	consoleOutput := widget.NewMultiLineEntry()
	consoleOutput.SetPlaceHolder("Console Output:")
	LoginUIOutputConsole = consoleOutput

	appendToConsole := func(text string) {
		currentText := LoginUIOutputConsole.Text
		LoginUIOutputConsole.SetText(fmt.Sprintf("%s\n%s", currentText, text))
	}

	var accountID int
	entry := widget.NewEntry()
	entry.SetPlaceHolder("ID of 1, 2, etc.")
	entry.Validator = func(s string) error {
		if len(s) > 2 {
			return fmt.Errorf("input too long")
		}
		return nil
	}

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Enter Email")
	pwEntry := widget.NewPasswordEntry()
	pwEntry.SetPlaceHolder("Enter Password")

	column1 := container.NewVBox(
		widget.NewLabel("Use Email and PW"),
		emailEntry,
		pwEntry,
		widget.NewButton("Login", func() {
			go func() {
				emailPwSelectedEnv := shared.GetCurrentEnvironment()
				email := emailEntry.Text
				pw := pwEntry.Text
				selectedEnvInt := uiFunctions.GetSelectIndex(shared.EnvOptions, shared.GetCurrentEnvironment())
				if email == "" || pw == "" {
					appendToConsole("Error: Email or Password is empty.")
					return
				}
				email, pw = EnterLoginCredentialsForGUI(selectedEnvInt, email, pw)
				appendToConsole(fmt.Sprintf("Logged into account with email %s in %s environment.", email, emailPwSelectedEnv))
				appendToConsole(fmt.Sprintf("Password: %s", pw))
			}()
		}),
	)

	column2 := container.NewVBox(
		widget.NewLabel(""), /* Spacer */
		widget.NewLabel("Enter Account ID:"),
		entry,
		widget.NewButton("Login", func() {
			go func() {
				selectedEnv := shared.GetCurrentEnvironment()
				enteredID := entry.Text
				selectedEnvInt := uiFunctions.GetSelectIndex(shared.EnvOptions, selectedEnv)
				verifyInt := testingToolkit.VerifyStringIsAnInt(enteredID)
				if testingToolkit.ConvertStringToInt(enteredID) > 21 {
					appendToConsole("Error: Account ID exceeds 21.")
					return
				}
				if len(enteredID) > 2 {
					appendToConsole("Error: Account ID exceeds 2 characters.")
					return
				}
				if verifyInt {
					accountID = testingToolkit.ConvertStringToInt(enteredID)
					email, pw := LoginToAccountForGUI(selectedEnvInt, accountID)
					appendToConsole(fmt.Sprintf("Logged into account %d in %s environment.", accountID, selectedEnv))
					appendToConsole(fmt.Sprintf("Email: %s", email))
					appendToConsole(fmt.Sprintf("Password: %s", pw))
				} else {
					appendToConsole("Error: Account ID must be a number.")
					return
				}
			}()
		}),
		widget.NewButton("Account Data", func() {
			go func() {
				selectedEnv := shared.GetCurrentEnvironment()
				enteredID := entry.Text
				selectedEnvInt := uiFunctions.GetSelectIndex(shared.EnvOptions, selectedEnv)
				verifyInt := testingToolkit.VerifyStringIsAnInt(enteredID)
				if testingToolkit.ConvertStringToInt(enteredID) > 21 {
					appendToConsole("Error: Account ID exceeds 21.")
					return
				}
				if len(enteredID) > 2 {
					appendToConsole("Error: Account ID exceeds 2 characters.")
					return
				}
				if verifyInt {
					accountID = testingToolkit.ConvertStringToInt(enteredID)
					accountData := runParametersForUI.GetAccountDataForGUI(selectedEnvInt)
					appendToConsole(fmt.Sprintf("%s Account:", selectedEnv))
					appendToConsole(fmt.Sprintf("Account ID: %d", accountID))
					appendToConsole(fmt.Sprintf("First Name: %s", accountData[accountID][1]))
					appendToConsole(fmt.Sprintf("Last Name: %s", accountData[accountID][2]))
					appendToConsole(fmt.Sprintf("Email: %s", accountData[accountID][3]))
					appendToConsole(fmt.Sprintf("Password: %s", accountData[accountID][4]))
					appendToConsole(fmt.Sprintf("JWT: %s", accountData[accountID][5]))
					appendToConsole(fmt.Sprintf("API Key ID: %s", accountData[accountID][6]))
					appendToConsole(fmt.Sprintf("Account UUID: %s", accountData[accountID][7]))
					appendToConsole(fmt.Sprintf("Gateway: %s", accountData[accountID][8]))
					appendToConsole(fmt.Sprintf("GatewayID: %s", accountData[accountID][9]))
				} else {
					appendToConsole("Error: Account ID must be a number.")
					return
				}
			}()
		}),
		widget.NewButton("Test Button", func() {
			go func() {
				appendToConsole("This is a test button. It does nothing. change 8")
			}()
		}),
		/*widget.NewButton("Secrets Only", func() {
			selectedEnv = envSelect.Selected
			accountData := runParametersForUI.GetAccountDataForGUI(uiFunctions.GetSelectIndex(envOptions, selectedEnv))
			for i := range accountData {
				appendToConsole(fmt.Sprintf("\"%s\",", accountData[i][4]))
				appendToConsole(fmt.Sprintf("\"%s\",", accountData[i][5]))
			}
		}),*/
	)

	bottomClearConsoleButton := widget.NewButton("     Clear Console     ", func() {
		go LoginUIOutputConsole.SetText("")
	})

	leftContent := container.NewVBox(
		column1,
		column2,
		layout.NewSpacer(),
		bottomClearConsoleButton,
	)

	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, LoginUIOutputConsole)

	return contentWithConsole

}
