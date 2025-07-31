package testRunParameters

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testAutomationSuiteGO/internal/projectData"
	"testAutomationSuiteGO/internal/testingToolkit"
)

var accountData = [][]string{}

func BuildAccountData(env string) [][]string {

	var accountsFolder string
	if strings.Contains(env, "1") {
		accountsFolder = testingToolkit.CurrPath() + "/config/accounts"
	} else {
		accountsFolder = testingToolkit.CurrPath() + "/config/githubActionsAccountFiles"
	}
	accountFilePath := accountsFolder + "/" //+ GetAccountJsonFile()
	if strings.Contains(accountFilePath, projectData.ProjectName+"/"+projectData.ProjectName) && strings.Contains(os.Args[1], "1") {
		accountFilePath = strings.ReplaceAll(accountFilePath, projectData.ProjectName+"/"+projectData.ProjectName, projectData.ProjectName)
	}
	var accountsInJson AccountDataRoot
	jsonFile, err := os.Open(accountFilePath)
	if err != nil {
		println(err)
	}
	defer jsonFile.Close()
	err = json.NewDecoder(jsonFile).Decode(&accountsInJson)
	if err != nil {
		println(err)
	}
	accountsArray := accountsInJson.AccountData
	for i := 0; i < len(accountsArray); i++ {
		account := accountsArray[i]
		if strings.Contains(env, "1") {
			accountData = append(accountData, []string{account.AccountInt, account.FirstName, account.LastName, account.Email, account.Password, account.Jwt, account.ApiKeyID, account.AccountId, account.Gateway, account.GatewayID})
		} else {
			accountsVariablesFromOS := ConvertOSArgs(os.Args)
			accountData = append(accountData, []string{account.AccountInt, account.FirstName, account.LastName, account.Email, accountsVariablesFromOS[i][0], accountsVariablesFromOS[i][1], account.ApiKeyID, account.AccountId, account.Gateway, account.GatewayID})
		}
	}
	return accountData
}

func BuildAccountDataV3(env string) [][]string {
	var accountsFolder string
	if strings.Contains(env, "1") {
		accountsFolder = testingToolkit.CurrPath() + "/config/accounts"
	} else {
		accountsFolder = testingToolkit.CurrPath() + "/config/githubActionsAccountFiles"
	}
	accountFilePath := accountsFolder + "/" //+ GetAccountJsonFile()
	var accountsInJson AccountDataRoot
	jsonFile, err := os.Open(accountFilePath)
	if err != nil {
		println(err)
	}
	defer jsonFile.Close()
	err = json.NewDecoder(jsonFile).Decode(&accountsInJson)
	if err != nil {
		println(err)
	}
	accountsArray := accountsInJson.AccountData
	for i := 0; i < len(accountsArray); i++ {
		account := accountsArray[i]
		if strings.Contains(env, "1") {
			accountData = append(accountData, []string{account.AccountInt, account.FirstName, account.LastName, account.Email, account.Password, account.Jwt, account.ApiKeyID, account.AccountId, account.Gateway, account.GatewayID})
		} else {
			accountsVariablesFromOS := ConvertOSArgsV3(os.Args[6])
			accountData = append(accountData, []string{account.AccountInt, account.FirstName, account.LastName, account.Email, accountsVariablesFromOS[i][0], accountsVariablesFromOS[i][1], account.ApiKeyID, account.AccountId, account.Gateway, account.GatewayID})
		}
	}
	return accountData
}

func ConvertOSArgs(args []string) [][]string {
	removed3FirstArgs := args[6:]
	rows := len(removed3FirstArgs) / 2
	convertedArgs := make([][]string, rows)
	for i := 0; i < rows; i++ {
		convertedArgs[i] = removed3FirstArgs[i*2 : i*2+2]
	}
	return convertedArgs
}

func ConvertOSArgsV3(args string) [][]string {
	var data SecretsData
	err := json.Unmarshal([]byte(args), &data)
	if err != nil {
		fmt.Printf("Error unmarshaling Secrets JSON: %v\n", err)
		return nil
	}

	theArgs := data.AccountSecrets

	rows := len(theArgs) / 2
	convertedArgs := make([][]string, rows)
	for i := 0; i < rows; i++ {
		convertedArgs[i] = theArgs[i*2 : i*2+2]
	}
	return convertedArgs
}

type SecretsData struct {
	AccountSecrets []string `json:"accountSecrets"`
}
