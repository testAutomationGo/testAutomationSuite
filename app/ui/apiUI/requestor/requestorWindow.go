package requestor

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const outputFolderLocation = "/testData/appOutput/apiRequestOutput"
const savedRequestsPreferenceKey = "storedRequests"

var storedRequests []SavedRequest

func setStoredRequests(requests []SavedRequest) {
	storedRequests = requests
}
func getStoredRequests() []SavedRequest {
	return storedRequests
}

var authSet bool = true
var authenticationType string
var authenticationToken string

func setAuthenticationType(tokenType string) {
	authenticationType = tokenType
}
func getAuthenticationType() string {
	return authenticationType
}

func setAuthenticationToken(token string) {
	authenticationToken = token
}
func getAuthenticationToken() string {
	return authenticationToken
}

func RequestorWindow(deps shared.AppDependencies) fyne.CanvasObject {
	if _, err := os.Stat(outputFolderLocation); os.IsNotExist(err) {
		os.MkdirAll(outputFolderLocation, 0755)
	}
	storedRequestsString := deps.App.Preferences().String(savedRequestsPreferenceKey)
	if storedRequestsString != "" {
		var requests []SavedRequest
		err := json.Unmarshal([]byte(storedRequestsString), &requests)
		if err != nil {
			uiFunctions.NotificationPopUp("Error loading stored requests.", "Error loading stored requests: "+err.Error(), deps)
		} else {
			setStoredRequests(requests)
		}
	}
	namesOfStoredRequests := make([]string, len(getStoredRequests()))
	for i, request := range getStoredRequests() {
		namesOfStoredRequests[i] = request.Name
	}
	consoleOutput := widget.NewMultiLineEntry()
	consoleOutput.SetPlaceHolder("Console Output")
	appendToConsole := func(text string) {
		fyne.Do(func() {
			currentText := consoleOutput.Text
			consoleOutput.SetText(fmt.Sprintf("%s\n%s", currentText, text))
		})
	}
	selectRequestTypeLabel := widget.NewLabel("Select Request Type:")
	requestTypeSelect := widget.NewSelect([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}, func(value string) {
	})
	requestTypeSelect.SetSelectedIndex(0)
	nameLabel := widget.NewLabel("Request Name:")
	nameEntry := widget.NewEntry()

	urlLabel := widget.NewLabel("URL:")
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://api.example.com/endpoint")

	printConsoleToFileButton := widget.NewButton("Print Console To File", func() {
		go func() {
			request := "Request: " + requestTypeSelect.Selected + " " + urlEntry.Text + "\n"
			auth := "Authentication: " + getAuthenticationType() + " " + getAuthenticationToken() + "\n"
			headers := "Headers: \n"
			testingToolkit.PrettyPrintJsonStringToFile(request+auth+headers+consoleOutput.Text, testingToolkit.CurrPath()+outputFolderLocation+"/currentRequestOutput_"+testingToolkit.CurrentTimeForNamingWithMS()+".txt")
		}()
	})
	openOutputFolderButton := widget.NewButton("Open Output Folder", func() {
		go uiFunctions.OpenFolder(testingToolkit.CurrPath() + outputFolderLocation)
	})
	bottomClearConsoleButton := widget.NewButton("Clear Console", func() {
		go consoleOutput.SetText("")
	})

	urlRow := container.New(layout.NewBorderLayout(nil, nil, urlLabel, nil), urlLabel, urlEntry)
	bodyTab := BodyTab(deps)
	controls := BuildHeaders(5)

	runRequestButton := widget.NewButton("Run Request", func() {
		go func() {
			fyne.Do(func() {
				consoleOutput.SetText("")
			})
			appendToConsole("Running...")
			if !ValidateURL(urlEntry.Text) {
				appendToConsole("Error: The URL format is invalid. Please correct the URL and try again.")
				return
			}
			var headers = make(map[string]string)
			bools := controls.Checks
			keys := controls.KeyEntries
			values := controls.ValueEntries
			headers = make(map[string]string)
			for i := range bools {
				if bools[i].Checked {
					headers[keys[i].Text] = values[i].Text
				}
			}
			if !ValidateHeaders(headers) {
				appendToConsole("Error: One or more headers are invalid. Please correct the headers and try again.")
				return
			}
			if bodyTab.Entry.Text != "" && !IsValidJSON(bodyTab.Entry.Text) {
				appendToConsole("Error: The JSON body format is invalid. Please correct the JSON format and try again.")
				return
			}
			RunApiRequest(appendToConsole, requestTypeSelect.Selected, urlEntry.Text, headers, TrimJson(bodyTab.Entry.Text), deps, false)
			go func() {
				newText := consoleOutput.Text
				newText = strings.ReplaceAll(consoleOutput.Text, "Running...", "Request Completed.")
				consoleOutput.SetText(newText)
			}()
		}()
	})
	runRequestButton.Importance = widget.HighImportance
	runRequestAndSaveBodyResponseToFileButton := widget.NewButton("Run and Save Response To File", func() {
		go func() {
			fyne.Do(func() {
				consoleOutput.SetText("")
			})
			appendToConsole("Running Request for: " + requestTypeSelect.Selected + " in the " + shared.GetCurrentEnvironment() + " environment.")
			if !ValidateURL(urlEntry.Text) {
				appendToConsole("Error: The URL format is invalid. Please correct the URL and try again.")
				return
			}
			var headers = make(map[string]string)
			bools := controls.Checks
			keys := controls.KeyEntries
			values := controls.ValueEntries
			headers = make(map[string]string)
			for i := range bools {
				if bools[i].Checked {
					headers[keys[i].Text] = values[i].Text
				}
			}
			if !ValidateHeaders(headers) {
				appendToConsole("Error: One or more headers are invalid. Please correct the headers and try again.")
				return
			}
			if bodyTab.Entry.Text != "" && !IsValidJSON(bodyTab.Entry.Text) {
				appendToConsole("Error: The JSON body format is invalid. Please correct the JSON format and try again.")
				return
			}
			RunApiRequest(appendToConsole, requestTypeSelect.Selected, urlEntry.Text, headers, TrimJson(bodyTab.Entry.Text), deps, true)
			go func() {
				newText := consoleOutput.Text
				newText = strings.ReplaceAll(consoleOutput.Text, "Running...", "Request Completed.")
				consoleOutput.SetText(newText)
			}()
		}()
	})
	setAuthButton := widget.NewButton("Set Authentication", func() {
		ShowAuthSetterDialog(deps)
	})

	buttonSaveRequest := widget.NewButton("Save Request", func() {
		go func() {
			SaveRequest(nameEntry.Text, urlEntry.Text, requestTypeSelect.Selected, nil, "", getAuthenticationType(), getAuthenticationToken(), deps)
		}()
	})

	storedRequestSelect := widget.NewSelect(namesOfStoredRequests, func(value string) {
		if len(namesOfStoredRequests) == 0 {
			return
		}
		for _, request := range getStoredRequests() {
			if request.Name == value {
				nameEntry.SetText(request.Name)
				urlEntry.SetText(request.URL)
				requestTypeSelect.SetSelected(request.RequestType)
				nameEntry.SetText(request.Name)
				if request.AuthenticationType != "" && request.AuthenticationToken != "" {
					authSet = true
				}
				setAuthenticationType(request.AuthenticationType)
				setAuthenticationToken(request.AuthenticationToken)
				break
			}
		}
	})
	topRow := container.New(layout.NewBorderLayout(nil, nil, nameLabel, nil), nameLabel, nameEntry)

	topContent := container.NewVBox(
		topRow,
		container.NewHBox(selectRequestTypeLabel, requestTypeSelect, layout.NewSpacer(), layout.NewSpacer(), runRequestAndSaveBodyResponseToFileButton, runRequestButton),
		urlRow,
		layout.NewSpacer(),
		container.NewHBox(setAuthButton, widget.NewLabel("Load Requests:"), storedRequestSelect, layout.NewSpacer(), printConsoleToFileButton, openOutputFolderButton, bottomClearConsoleButton, buttonSaveRequest),
	)
	tabs := []ApiLowerTab{
		{
			name:    "Console Output",
			content: consoleOutput,
		},
		{
			name:    "Body",
			content: bodyTab.Container,
		},
		{
			name:    "Headers",
			content: controls.Container,
		},
		{
			name:    "Params",
			content: ParamsTab(),
		},
	}
	tabItems := make([]*container.TabItem, len(tabs))
	for i, tab := range tabs {
		tabItems[i] = container.NewTabItem(tab.name, tab.content)
	}
	apiLowerContent := container.NewAppTabs(tabItems...)
	split := container.NewVSplit(topContent, apiLowerContent)
	split.SetOffset(0.18)
	return split
}

type ApiLowerTab struct {
	name    string
	content fyne.CanvasObject
}

type BodyContent struct {
	Container *fyne.Container
	Entry     *widget.Entry
}

func BodyTab(deps shared.AppDependencies) BodyContent {
	jsonBodyLabel := widget.NewLabel("JSON Body:")
	jsonBodyEntry := widget.NewMultiLineEntry()
	jsonBodyEntry.SetText(`{
	  "key": "value",
	  "set": "data"
	}`)
	checkJsonBodyFormatButton := widget.NewButton("Check JSON Format", func() {
		jsonStr := jsonBodyEntry.Text
		if jsonStr == "" {
			uiFunctions.NotificationPopUp("Empty JSON", "The JSON body is emptyd.", deps)
			return
		}
		isValid := IsValidJSON(jsonStr)
		if isValid {
			uiFunctions.NotificationPopUp("Valid JSON", "The JSON format is valid.", deps)
		} else {
			uiFunctions.NotificationPopUp("Error: Invalid JSON", "Error: The JSON format is invalid.", deps)
		}
	})
	clearJsonBodyEntryButton := widget.NewButton("Clear Body", func() {
		go jsonBodyEntry.SetText("")
	})
	topRow := container.NewHBox(
		jsonBodyLabel,
		layout.NewSpacer(),
		layout.NewSpacer(),
		clearJsonBodyEntryButton,
		checkJsonBodyFormatButton,
	)
	bodyContent := container.New(
		layout.NewBorderLayout(topRow, nil, nil, nil),
		topRow,
		jsonBodyEntry,
	)
	return BodyContent{
		Container: bodyContent,
		Entry:     jsonBodyEntry,
	}
}

type Controls struct {
	Container    *fyne.Container
	Checks       []*widget.Check
	KeyEntries   []*widget.Entry
	ValueEntries []*widget.Entry
}

func BuildHeaders(rows int) *Controls {
	headersLabel := widget.NewLabel("Headers:")
	var checks []*widget.Check
	var keyEntries []*widget.Entry
	var valueEntries []*widget.Entry
	var lines []fyne.CanvasObject
	for i := range rows {
		chk := widget.NewCheck(fmt.Sprintf("%d", i+1), nil)
		key := widget.NewEntry()
		val := widget.NewEntry()

		entries := container.NewAdaptiveGrid(2, key, val)
		colon := widget.NewLabel(":")
		centerOverlay := container.NewStack(entries, container.NewCenter(colon))
		line := container.New(
			layout.NewBorderLayout(nil, nil, chk, nil),
			chk,
			centerOverlay,
		)
		checks = append(checks, chk)
		keyEntries = append(keyEntries, key)
		valueEntries = append(valueEntries, val)
		lines = append(lines, line)
	}
	content := container.NewVBox(headersLabel, container.NewVBox(lines...))
	return &Controls{
		Container:    content,
		Checks:       checks,
		KeyEntries:   keyEntries,
		ValueEntries: valueEntries,
	}
}

func ParamsTab() fyne.CanvasObject {
	paramsLabel := widget.NewLabel("Params:")
	return paramsLabel
}
