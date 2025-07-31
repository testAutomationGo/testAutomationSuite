package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testAutomationSuiteGO/app/mainFrame"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/ui/runParametersForUI"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

//Deploy Command: go build -ldflags "-H=windowsgui" -o TestingApp.exe appMain.go

type CustomApp struct {
	fyne.App
}

func NewCustomApp() *CustomApp {
	a := app.NewWithID("com.example.TestingApp")
	return &CustomApp{App: a}
}

func main() {
	todaysDate := testingToolkit.CurrentDate()
	todaysDate = strings.ReplaceAll(todaysDate, "-", "")
	logFolder := "app/logs"
	if _, err := os.Stat(logFolder); os.IsNotExist(err) {
		os.Mkdir(logFolder, 0755)
	}
	file, err := os.OpenFile("app/logs/app"+todaysDate+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("Failed to open log file:", err)
	}
	log.SetOutput(file)
	log.Println("Launching Testing App")
	TestingApp()
	log.Println("Closing Testing App")
}

func TestingApp() {
	runParametersForUI.SetGUIConfiguration()
	shared.SetEnvOptions()
	uiFunctions.CreateAppFolderOutput()

	myApp := NewCustomApp()
	myWindow := mainFrame.AppWindow(myApp)
	mainFrame.SetAppIcon(myWindow)

	shared.SetThisApp(myApp)
	shared.SetThisWindow(myWindow)

	dependencies := shared.AppDependencies{
		App:        shared.GetThisApp(),
		MainWindow: shared.GetThisWindow(),
	}

	// Example of storing a credential
	myApp.Preferences().SetString("username", "exampleUser")

	// Example of retrieving a credential
	username := myApp.Preferences().String("username")
	fmt.Println("Stored username:", username)

	mainFrame.SetupTabs(dependencies)
}
