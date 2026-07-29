package shared

import (
	"encoding/json"
	"log"
	"os"

	"fyne.io/fyne/v2"
)

type AppDependencies struct {
	App        fyne.App
	MainWindow fyne.Window
}

const (
	ButtonWidth  = 200
	ButtonHeight = 50
)

var thisApp fyne.App

func SetThisApp(app fyne.App) {
	thisApp = app
}

func GetThisApp() fyne.App {
	return thisApp
}

var thisWindow fyne.Window

func SetThisWindow(window fyne.Window) {
	thisWindow = window
}
func GetThisWindow() fyne.Window {
	return thisWindow
}

var currentEnvironment string

var EnvOptions []string

var TestRunEnvVariables []string

var appConfigFromFile AppConfiguration

func SetEnvOptions() {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Println("Error getting current working directory:", err)
		return
	}
	appConfigFile := workingDir + "/config/appConfig.json"
	jsonFile, err := os.Open(appConfigFile)
	if err != nil {
		log.Println("Error opening appConfig.json:", err)
		return
	}
	defer jsonFile.Close()
	err = json.NewDecoder(jsonFile).Decode(&appConfigFromFile)
	if err != nil {
		log.Println("Error decoding appConfig.json:", err)
		return
	}
	EnvOptions = appConfigFromFile.Options
	if len(EnvOptions) == 0 {
		log.Println("No environment options found in appConfig.json")
		return
	}
	log.Println("Environment options loaded:", appConfigFromFile.Envs[2])
	SetCurrentEnvironment(appConfigFromFile.Options[2])
}

func GetAppConfigFromFile() AppConfiguration {
	return appConfigFromFile
}

type AppConfiguration struct {
	Envs                  []string `json:"environments"`
	Options               []string `json:"envOptions"`
	RunTestsEnvVariables  []string `json:"runTestsEnvVariables"`
	TestSectorNames       []string `json:"testSectorNames"`
	EnvConfigurationFiles []string `json:"envConfigurationFiles"`
}

func SetCurrentEnvironment(env string) {
	convertedEnv := EnvConverterForRunVariable(env)
	log.Println("Setting current environment to:", convertedEnv)
	currentEnvironment = convertedEnv
}

func GetCurrentEnvironment() string {
	return currentEnvironment
}

func EnvConverterForRunVariable(envString string) string {
	var envIntCheck int
	envOptions := EnvOptions
	for i := range envOptions {
		if envOptions[i] == envString {
			envIntCheck = i
			break
		}
	}
	envVariables := TestRunEnvVariables
	return envVariables[envIntCheck]
}

func GetCurrentEnvInt() int {
	for i := range TestRunEnvVariables {
		if TestRunEnvVariables[i] == GetCurrentEnvironment() {
			return i
		}
	}
	return 0
}

var currentUserJWT string

func SetCurrentUserJWT(jwt string) {
	currentUserJWT = jwt
}

func GetCurrentUserJWT() string {
	if currentUserJWT == "" {
		return "No JWT set"
	}
	return currentUserJWT
}

var useUserJWT bool

func SetUseUserJWT(use bool) {
	useUserJWT = use
}

func GetUseUserJWT() bool {
	return useUserJWT
}
