package mainFrame

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/ui/runParametersForUI"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type Tab struct {
	name    string
	content fyne.CanvasObject
}

func SetAppIcon(myWindow fyne.Window) {
	iconPath := testingToolkit.CurrPath() + "/app/assets/cyclesSquareImage.png"
	iconResource, err := fyne.LoadResourceFromPath(iconPath)
	if err != nil {
		log.Println(iconPath)
		log.Println("Could not load icon resource:", err)
	} else {
		myWindow.SetIcon(iconResource)
	}
}

func OpenFolder(folderPath string) error {

	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err := os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			log.Println("Could not create folder:", err)
			return fmt.Errorf("failed to create folder: %v", err)
		}
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", strings.ReplaceAll(folderPath, "/", "\\"))
	case "darwin":
		cmd = exec.Command("open", folderPath)
	case "linux":
		cmd = exec.Command("xdg-open", folderPath)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

func OpenFile(filePath string) error {

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Println("File does not exist:", err)
		return fmt.Errorf("file does not exist: %v", err)
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", strings.ReplaceAll(filePath, "/", "\\"))
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "linux":
		cmd = exec.Command("xdg-open", filePath)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

func ShowFindDialog(win fyne.Window) {
	consoleOutput := runParametersForUI.GetGUIConsoleOutput()
	entry := widget.NewEntry()
	findButton := widget.NewButton("Find", func() {
		searchText := entry.Text
		content := consoleOutput.Text
		highlightedContent := uiFunctions.HighlightText(content, searchText)
		consoleOutput.SetText(highlightedContent)
	})

	dialog := dialog.NewCustom("Find", "Close", container.NewVBox(
		widget.NewLabel("Enter text to find, will be highlighted by ===== =====:"),
		entry,
		findButton,
	), win)

	dialog.Show()
}

func ShowPreferencesDialog(deps shared.AppDependencies) {
	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Enter Email")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter Password")

	jwtEntry := widget.NewPasswordEntry()
	jwtEntry.SetPlaceHolder("Enter JWT")

	storedEmail := deps.App.Preferences().String("email")
	storedPassword := deps.App.Preferences().String("password")
	storedJWT := deps.App.Preferences().String("jwt")
	emailEntry.SetText(storedEmail)
	passwordEntry.SetText(storedPassword)
	jwtEntry.SetText(storedJWT)

	widget.NewFormItem("Email", emailEntry)
	widget.NewFormItem("Password", passwordEntry)
	widget.NewFormItem("JWT", jwtEntry)

	formItems := []*widget.FormItem{
		widget.NewFormItem("Email", emailEntry),
		widget.NewFormItem("Password", passwordEntry),
		widget.NewFormItem("JWT", jwtEntry),
	}

	form := widget.NewForm(formItems...)

	dialogContent := container.NewVBox(form)
	dialogContent.Resize(fyne.NewSize(400, 250))

	d := dialog.NewCustomConfirm("Preferences", "Save", "Cancel", dialogContent, func(b bool) {
		if b {
			if !(emailEntry.Text == "") {
				deps.App.Preferences().SetString("email", emailEntry.Text)
				deps.App.Preferences().SetString("userName", strings.Split(emailEntry.Text, "@")[0])
			}
			if !(passwordEntry.Text == "") {
				deps.App.Preferences().SetString("password", passwordEntry.Text)
			}
			if !(jwtEntry.Text == "") {
				deps.App.Preferences().SetString("jwt", jwtEntry.Text)
			}
		}
	}, deps.MainWindow)

	d.Resize(fyne.NewSize(400, 250))
	d.Show()
}
