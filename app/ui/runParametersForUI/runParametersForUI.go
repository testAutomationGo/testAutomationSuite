package runParametersForUI

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/testRunParameters"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var envOptions []string

var environments []string

var envInts []int

var testSectorNames []string

var testSectorCamelCase []string

var envConfigFiles []string

func SetGUIConfiguration() {
	workingDir := testingToolkit.CurrPath()
	configFiles, err := uiFunctions.ListFilesInFolder(workingDir + "/config")
	if err != nil {
		log.Println("No config files found in the config folder.")
		fmt.Println("No config files found in the config folder.")
		return
	}
	var appConfig AppConfiguration
	for _, file := range configFiles {
		if strings.Contains(file, "appConfig") {
			configFilePath := workingDir + "/config/" + file
			err := uiFunctions.ReadJSONFile(configFilePath, &appConfig)
			if err != nil {
				log.Println("Could not read appConfig file:", err)
				println(err)
			}
		}
	}
	envOptions = appConfig.Options
	environments = appConfig.Envs
	testSectorNames = appConfig.TestSectorNames
	testSectorCamelCase = appConfig.TestSectorCamelCase
	envConfigFiles = appConfig.EnvConfigurationFiles
	numOEnvs := len(environments)
	for i := 0; i < numOEnvs; i++ {
		envInts = append(envInts, i)
	}
}

func GetEnvOptions() []string {
	return envOptions
}

func GetEnvironments() []string {
	return environments
}

func GetEnvInts() []int {
	return envInts
}

func GetTestSectorNames() []string {
	return testSectorNames
}

func GetTestSectorCamelCase() []string {
	return testSectorCamelCase
}

func GetEnvConfigFiles() []string {
	return envConfigFiles
}

func GetWebAppUrl(envInt int) string {
	envConFiles := GetEnvConfigFiles()
	envFile := envConFiles[envInt]
	var configRunnerRoot testRunParameters.ConfigRunnerRoot
	jsonFile, err := os.Open(testingToolkit.CurrPath() + "/config/" + envFile)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}
	defer jsonFile.Close()
	err = json.NewDecoder(jsonFile).Decode(&configRunnerRoot)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}
	return configRunnerRoot.Env.AppUrl
}

func GetAccountDataForGUI(envInt int) [][]string {
	var accounts [][]string
	envConFiles := GetEnvConfigFiles()
	envFile := envConFiles[envInt]
	var configRunnerRoot testRunParameters.ConfigRunnerRoot
	jsonFile, err := os.Open(testingToolkit.CurrPath() + "/config/" + envFile)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}
	defer jsonFile.Close()
	err = json.NewDecoder(jsonFile).Decode(&configRunnerRoot)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}
	accountFile := ""
	var accountsRoot AccountsRoot
	jsonFile, err = os.Open(testingToolkit.CurrPath() + "/config/accounts/" + accountFile)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}
	defer jsonFile.Close()
	err = json.NewDecoder(jsonFile).Decode(&accountsRoot)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}
	accountsArray := accountsRoot.AccountData
	for i := 0; i < len(accountsArray); i++ {
		account := accountsArray[i]
		accounts = append(accounts, []string{account.AccountInt, account.FirstName, account.LastName, account.Email, account.Password, account.Jwt, account.ApiKeyID, account.AccountID, account.Gateway, account.GatewayID})
	}
	return accounts
}

var (
	GUIConsoleOutput *widget.Entry
	mu               sync.Mutex
)

func SetGUIConsoleOutput() {
	consoleOutput := widget.NewMultiLineEntry()
	consoleOutput.SetPlaceHolder("Console Output:")
	consoleOutput.Wrapping = fyne.TextWrapWord
	GUIConsoleOutput = consoleOutput
}

func GetGUIConsoleOutput() *widget.Entry {
	textToSet := GUIConsoleOutput.Text
	GUIConsoleOutput.SetText(textToSet)
	return GUIConsoleOutput
}

func UpdateConsoleOutput(text string) {
	mu.Lock()
	defer mu.Unlock()
	consoleOutput := GetGUIConsoleOutput()
	consoleOutput.SetText(consoleOutput.Text + "\n" + text)
}

func ClearConsoleOutput() {
	mu.Lock()
	defer mu.Unlock()
	consoleOutput := GetGUIConsoleOutput()
	consoleOutput.SetText("")
}

var GuiRun = false

func SetGuiRun() {
	GuiRun = true
}

func GetGuiRun() bool {
	return GuiRun
}

func SetGUIEnvConfigurationsForAppFunctions(env string) {
	testRunParameters.SetConfigFile(env + "1")
	testRunParameters.SetConfigRunnerJsonParameters()
	testRunParameters.BuildAccountData(env + "1")
}

func GetEnvString(value string) string {
	var n int
	for i, v := range envOptions {
		if v == value {
			n = i
		}
	}
	return environments[n]
}

func GetEnvStringFromInt(value int) string {
	return environments[value]
}

var DBNames []string

var ConnectionStrings []string

func SetDBNamesAndConnectionStrings() {
	configFile := testingToolkit.CurrPath() + "config/uiArgs/uiArgs.json"
	var uiArgsRoot UIArgsRoot
	err := uiFunctions.ReadJSONFile(configFile, &uiArgsRoot)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}
	dbConnections := uiArgsRoot.UIArgs.DatabaseConnectionStrings
	for i := 0; i < len(dbConnections); i++ {
		if dbConnections[i].DBTables == nil {
			continue
		}
		if len(dbConnections[i].DBTables) == 0 {
			continue
		}
		dbConnection := dbConnections[i]
		DBNames = append(DBNames, dbConnection.DatabaseName)
		ConnectionStrings = append(ConnectionStrings, dbConnection.ConnectionString)
	}
}

func GetDBNames() []string {
	return DBNames
}

func GetConnectionStrings() []string {
	return ConnectionStrings
}
