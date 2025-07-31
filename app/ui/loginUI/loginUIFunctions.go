package loginUI

import (
	"testAutomationSuiteGO/app/shared"
)

func LoginToAccountForGUI(envInt int, accountID int) (string, string) {
	email := ""
	password := ""
	return email, password
}

func EnterLoginCredentialsForGUI(envInt int, email, password string) (string, string) {
	return email, password
}

func SelectAccountIDForLogin(envInt int, deps shared.AppDependencies) int {
	accountID := 0
	return accountID
}
