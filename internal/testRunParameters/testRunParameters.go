package testRunParameters

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testAutomationSuiteGO/app/uiFunctions"
	testingToolkit "testAutomationSuiteGO/internal/testingToolkit"
)

var executionStartTime string

var configRunnerFile string
var configRunner ConfigRunnerRoot

// Config file variables, set in the environment configuration files.
var envName string
var envInt int
var envString string
var evnLowers string
var envUppers string
var apiEndpoint string
var appUrl string
var mobileAppID string
var localVariableFile string

var runLocal = false
var runID string

//var SMEEURL string

var finalResultsFolderName = "defaultFolderName"

func SetExecutionStartTime() {
	executionStartTime = testingToolkit.ReadableCurrentTimeWithMS()
}

func GetExecutionStartTime() string {
	return executionStartTime
}

func SetConfigFile(env string) {
	//Ensure specificity when running a test suite locally or on a local machine.
	//If creating a local env then name config file with lower case "l", ex: localConfigRunner.json. This
	//is to avoid confusion with the "Local" at the end of the file name, in which this "Local"
	//means the tests will run "on a machine locally with a headless browser" not local env(https://localhost).
	//All config files must have the env name at the beginning and the first argument for the execution will need to match the env name.
	//args in the run command: go run main.go envName gmailCreds notionToken slackToken
	//if using the app ensure to fill out the args in the argsAfterEnv.json file.
	if strings.Contains(env, "1") {
		env = strings.ReplaceAll(env, "1", "")
		runLocal = true
	}
	configFolder := testingToolkit.CurrPath() + "/config"
	files := testingToolkit.ListFilesInFolder(configFolder)
	for _, file := range files {
		if strings.Contains(file, env) {
			configRunnerFile = file
		}
	}
	log.Println("Config File: " + configRunnerFile)
}

func GetFullConfigFilePath() string {
	path := testingToolkit.CurrPath() + "/config/" + configRunnerFile
	return path
}

func SetConfigRunnerJsonParameters() {
	jsonFile, err := os.Open(GetFullConfigFilePath())
	if err != nil {
		fmt.Println("Error opening config runner file:", err)
	}
	defer jsonFile.Close()
	err = json.NewDecoder(jsonFile).Decode(&configRunner)
	if err != nil {
		fmt.Println("Error decoding config runner JSON:", err)
	}
	envName = configRunner.Env.Name
	envInt = configRunner.Env.EnvInt
	envString = configRunner.Env.EnvString
	evnLowers = configRunner.Env.EnvLowers
	envUppers = configRunner.Env.EnvUppers
	apiEndpoint = configRunner.Env.ApiEndpoint
	appUrl = configRunner.Env.AppUrl
	mobileAppID = configRunner.Env.MobileAppID
	localVariableFile = configRunner.Env.LocalVariableFile
}

func GetConfigRunner() ConfigRunnerRoot {
	return configRunner
}

func GetEnvName() string {
	return envName
}

func GetEnvInt() int {
	return envInt
}

func GetEnvString() string {
	return envString
}

func GetEnvLowers() string {
	return evnLowers
}

func GetEnvUppers() string {
	return envUppers
}

func GetApiEndpoint() string {
	return apiEndpoint
}

func GetWebAppUrl() string {
	return appUrl
}

func GetMobileAppID() string {
	return mobileAppID
}

func GetLocalVariableFile() string {
	return localVariableFile
}

func GetRunLocal() bool {
	return runLocal
}

func SetResultsFolderName() {
	runTime := testingToolkit.CurrentTimeForNamingWithMS()
	folderName := GetEnvUppers() + "_" + runTime
	folderLocation := testingToolkit.CurrPath() + "/resultsArchive/"
	folderPath := folderLocation + folderName
	err := os.MkdirAll(folderPath, 0755)
	if err != nil {
		fmt.Println("Error creating results folder:", err)
	}
	err = os.MkdirAll(folderPath+"/maestroFlows", 0755)
	if err != nil {
		fmt.Println("Error creating maestroFlows folder:", err)
	}
	finalResultsFolderName = folderName
}

func GetResultsFolderName() string {
	return finalResultsFolderName
}

func GetResultsFolderPath() string {
	path := testingToolkit.CurrPath() + "/resultsArchive/" + GetResultsFolderName()
	return path
}

type argsAfterEnv struct {
	ArgsAfterEnv []string `json:"args"`
}

func GetArgsAfterEnv() []string {
	jsonFile, err := os.Open(testingToolkit.CurrPath() + "/config/argsAfterEnv.json")
	if err != nil {
		fmt.Println("Error opening argsAfterEnv file:", err)
	}
	defer jsonFile.Close()
	var args argsAfterEnv
	err = json.NewDecoder(jsonFile).Decode(&args)
	if err != nil {
		fmt.Println("Error decoding argsAfterEnv JSON:", err)
	}
	return args.ArgsAfterEnv
}

// Current configuration is that the gmail credential is the first argument after the env name and the notion token is the second argument after the env name.
func GetGmailToken() string {
	if GetRunLocal() {
		return GetGmailCredsForUI()
	}
	if len(GetArgsAfterEnv()) < 1 {
		return "noToken"
	}
	if len(os.Args) < 3 {
		return GetArgsAfterEnv()[0]
	}
	gmailCreds := os.Args[2]
	return gmailCreds
}

func GetNotionToken() string {
	if len(GetArgsAfterEnv()) < 2 {
		return "noToken"
	}
	if len(os.Args) < 4 {
		return GetArgsAfterEnv()[1]
	}
	notionToken := os.Args[3]
	return notionToken
}

func GetRunID() string {
	if len(os.Args) > 4 {
		runID = os.Args[5]
	} else {
		runID = "No Run ID"
	}
	return runID
}

type UiArgs struct {
	Args `json:"uiArgs"`
}

type Args struct {
	GmailCreds string `json:"gmailCreds"`
}

func GetGmailCredsForUI() string {
	var uiArgs UiArgs
	err := uiFunctions.ReadJSONFile(testingToolkit.CurrPath()+"/config/uiArgs/uiArgs.json", &uiArgs)
	if err != nil {
		log.Println("Error reading uiArgs.json:", err)
		fmt.Println("Error reading uiArgs.json:", err)
	}
	return uiArgs.GmailCreds
}
