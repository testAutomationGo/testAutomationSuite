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
	TestRunEnvVariables = appConfigFromFile.Envs
	if len(TestRunEnvVariables) != len(EnvOptions) {
		log.Println("appConfig.json: environments and envOptions must hold the same number of entries")
		return
	}
	startupEnv := defaultEnvIndex()
	log.Println("Environment options loaded, starting in:", TestRunEnvVariables[startupEnv])
	SetCurrentEnvironment(EnvOptions[startupEnv])
}

// defaultEnvIndex resolves the environment the app opens in from the
// defaultEnvironment key in appConfig.json, falling back to the last
// environment listed. Callers must ensure Envs is non-empty.
func defaultEnvIndex() int {
	for i := range appConfigFromFile.Envs {
		if appConfigFromFile.Envs[i] == appConfigFromFile.DefaultEnvironment {
			return i
		}
	}
	log.Println("defaultEnvironment not matched in environments, using:", appConfigFromFile.Envs[len(appConfigFromFile.Envs)-1])
	return len(appConfigFromFile.Envs) - 1
}

// GetDefaultEnvOption returns the env option code the app opens in, ex: PRD.
func GetDefaultEnvOption() string {
	if len(EnvOptions) == 0 || len(appConfigFromFile.Envs) == 0 {
		log.Println("No environment options loaded, cannot resolve the startup environment")
		return ""
	}
	return EnvOptions[defaultEnvIndex()]
}

func GetAppConfigFromFile() AppConfiguration {
	return appConfigFromFile
}

type AppConfiguration struct {
	Envs                  []string `json:"environments"`
	Options               []string `json:"envOptions"`
	DefaultEnvironment    string   `json:"defaultEnvironment"`
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
	envOptions := EnvOptions
	envVariables := TestRunEnvVariables
	for i := range envOptions {
		if envOptions[i] == envString && i < len(envVariables) {
			return envVariables[i]
		}
	}
	log.Println("No environment matches the option:", envString)
	return ""
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
