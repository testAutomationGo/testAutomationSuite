package uiFunctions

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ListFilesInFolder(folderPath string) ([]string, error) {
	dir, err := os.Open(folderPath)
	if err != nil {
		log.Println("Error opening folder: ", err)
		return nil, err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		log.Println("Error reading folder: ", err)
		return nil, err
	}
	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}
	return fileNames, nil
}

func ReadJSONFile(filePath string, v interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening file: ", err)
		return err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		log.Println("Error reading file: ", err)
		return err
	}

	return json.Unmarshal(byteValue, v)
}

func OpenFolder(folderPath string) error {

	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err := os.MkdirAll(folderPath, os.ModePerm)
		if err != nil {
			log.Println("Error creating folder: ", err)
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
		log.Println("File does not exist: ", err)
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

func CreateAppFolderOutput() {
	currDir := testingToolkit.CurrPath()
	if strings.Contains(currDir, "/app") || strings.Contains(currDir, "\\app") {
		currDir = strings.ReplaceAll(currDir, "/app", "")
		currDir = strings.ReplaceAll(currDir, "\\app", "")
	}
	testDataFolder := currDir + "/testData"
	if _, err := os.Stat(testDataFolder); os.IsNotExist(err) {
		err := os.MkdirAll(testDataFolder, os.ModePerm)
		if err != nil {
			log.Println("Error creating folder: ", err)
			return
		}
	}
	appOutputFolder := currDir + "/testData/appOutput"
	if _, err := os.Stat(appOutputFolder); os.IsNotExist(err) {
		err := os.MkdirAll(appOutputFolder, os.ModePerm)
		if err != nil {
			log.Println("Error creating folder: ", err)
			return
		}
	}
}

func NotificationPopUp(title string, content string, deps shared.AppDependencies) {
	go func() {
		d := dialog.NewCustom(title, "Close", container.NewVBox(
			widget.NewLabel(content),
		), deps.MainWindow)
		d.Show()
	}()
}

func HighlightText(content, searchText string) string {
	if searchText == "" {
		return content
	}

	highlightStart := "===="
	highlightEnd := "===="

	highlightedContent := strings.ReplaceAll(content, searchText, highlightStart+searchText+highlightEnd)
	return highlightedContent
}

func AppendToGUIConsole(consoleOutput *widget.Entry, text string) {
	currentText := consoleOutput.Text
	consoleOutput.SetText(fmt.Sprintf("%s\n%s", currentText, text))
}

func UpdateGUIConsole(consoleOutput *widget.Entry, text string) {
	consoleOutput.SetText(text)
}

func GetSelectIndex(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

func UpdateConsoleForJsonPrettierPrint(consoleOutput *widget.Entry, text string) {
	prettyText := PrettyPrintJSON(text)
	consoleOutput.SetText(prettyText)
}

func AppendConsoleForJsonPrettierPrint(consoleOutput *widget.Entry, text string) {
	currentText := consoleOutput.Text
	prettyText := PrettyPrintJSON(text)
	consoleOutput.SetText(fmt.Sprintf("%s\n%s", currentText, prettyText))
}

func PrettyPrintJSON(jsonString string) string {
	var result strings.Builder
	indentLevel := 0
	inQuotes := false

	for i := 0; i < len(jsonString); i++ {
		char := jsonString[i]
		switch char {
		case '{', '[':
			if !inQuotes {
				result.WriteByte(char)
				result.WriteByte('\n')
				indentLevel++
				result.WriteString(strings.Repeat("    ", indentLevel))
			} else {
				result.WriteByte(char)
			}
		case '}', ']':
			if !inQuotes {
				result.WriteByte('\n')
				indentLevel--
				result.WriteString(strings.Repeat("    ", indentLevel))
				result.WriteByte(char)
			} else {
				result.WriteByte(char)
			}
		case ',':
			result.WriteByte(char)
			if !inQuotes {
				result.WriteByte('\n')
				result.WriteString(strings.Repeat("    ", indentLevel))
			}
		case ':':
			result.WriteByte(char)
			if !inQuotes {
				result.WriteByte(' ')
			}
		case '"':
			result.WriteByte(char)
			if i > 0 && jsonString[i-1] != '\\' {
				inQuotes = !inQuotes
			}
		default:
			result.WriteByte(char)
		}
	}

	return result.String()
}

func NewCustomEntry(window fyne.Window) *CustomEntry {
	entry := &CustomEntry{window: window}
	entry.ExtendBaseWidget(entry)
	return entry
}

type CustomEntry struct {
	widget.Entry
	window fyne.Window
}
