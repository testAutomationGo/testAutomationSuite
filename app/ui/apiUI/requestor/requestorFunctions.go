package requestor

import (
	"encoding/json"
	"fmt"
	"mime"
	"os"
	"strings"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/apiFunctions"
	"testAutomationSuiteGO/internal/testingToolkit"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func RunApiRequest(appendToConsole func(string), requestType string, url string, headers map[string]string, body string, deps shared.AppDependencies, saveToFile bool) {
	if authSet {
		if authenticationType != "" && authenticationToken != "" {
			headers["Authorization"] = authenticationType + " " + authenticationToken
			headers["Accept"] = "*/*"
		} else {
			appendToConsole("Error: Authentication type or token not set.")
			return
		}
	}
	jsonBody := make(map[string]any)
	var bodyKind apiFunctions.BodyKind
	if body != "" {
		bodyKind = apiFunctions.BodyJSON
		err := json.Unmarshal([]byte(body), &jsonBody)
		if err != nil {
			appendToConsole("Error parsing JSON body: " + err.Error())
			return
		}
	} else {
		bodyKind = apiFunctions.BodyNone
	}
	code, response, responseHeaders, err := apiFunctions.DoRequest(apiFunctions.RequestOptions{
		Method:   requestType,
		URL:      url,
		Headers:  headers,
		BodyKind: bodyKind,
		JSONBody: jsonBody,
		Timeout:  30 * time.Second,
	})
	if err != nil {
		appendToConsole("Error running api request: " + err.Error())
		return
	}
	if code != 200 && code != 201 {
		appendToConsole("Error running api request. Response Code: " + fmt.Sprintf("%d", code) + "\nResponse: " + string(response))
		return
	}
	contentDisposition := getHeaderValue(responseHeaders, "Content-Disposition")
	if strings.Contains(contentDisposition, "attachment") {
		fileName := getFileNameFromHeaders(responseHeaders)
		if fileName == "" {
			extension := inferExtentionFromHeaders(responseHeaders)
			fileName = "response" + extension
		}
		appendToConsole("Attachment detected. File name: " + fileName)
		outFile := testingToolkit.CurrPath() + outputFolderLocation + "/" + fileName
		err := os.WriteFile(outFile, response, 0644)
		if err != nil {
			appendToConsole("Error saving attachment to file: " + err.Error())
			return
		}
		appendToConsole("Saved To File In Request Response Output Folder: " + fileName)
		return
	}
	if !saveToFile {
		if len(response) > 100000 {
			go func() {
				testingToolkit.PrettyPrintJsonStringToFile(string(response), testingToolkit.CurrPath()+outputFolderLocation+"/requestOutput_"+testingToolkit.CurrentTimeForNamingWithMS()+".txt")
				appendToConsole("Response Code: " + fmt.Sprintf("%d", code))
				appendToConsole("Large response saved to file.")
			}()
			return
		} else {
			testingToolkit.PrettyPrintJsonStringToFile(string(response), testingToolkit.CurrPath()+outputFolderLocation+"/requestOutput_"+testingToolkit.CurrentTimeForNamingWithMS()+".txt")
			appendToConsole("Saved To File In Request Response Output Folder: " + fmt.Sprintf("%d", code))
		}

	}
}

func getFileNameFromHeaders(headers map[string][]string) string {
	contentDisposition := getHeaderValue(headers, "Content-Disposition")
	if contentDisposition == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(contentDisposition)
	if err != nil {
		return ""
	}
	if fileName, ok := params["filename"]; ok {
		return strings.Trim(fileName, "\" ")
	}
	return ""
}

func inferExtentionFromHeaders(headers map[string][]string) string {
	contentType := getHeaderValue(headers, "Content-Type")
	if contentType == "" {
		return ".bin"
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return ".bin"
	}
	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil || len(extensions) == 0 {
		return ".bin"
	}
	return extensions[0]
}

func getHeaderValue(headers map[string][]string, key string) string {
	for k, values := range headers {
		if strings.EqualFold(k, key) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func ValidateURL(url string) bool {
	return strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://")
}

func IsValidJSON(input string) bool {
	input = TrimJson(input)
	if input == "" {
		return false
	}
	return json.Valid([]byte(input))
}

func TrimJson(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "\n", "")
	input = strings.ReplaceAll(input, "\r", "")
	input = strings.ReplaceAll(input, "\t", "")
	input = strings.ReplaceAll(input, " ", "")
	return input
}

func ValidateHeaders(headers map[string]string) bool {
	for k, v := range headers {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			return false
		}
	}
	return true
}

func ShowAuthSetterDialog(deps shared.AppDependencies) {
	bearerCheckbox := widget.NewCheck("Use Bearer Token Authentication", func(checked bool) {
		if checked {
			setAuthenticationType("Bearer")
		} else {
			setAuthenticationType("")
		}
	})
	basicAuthCheckbox := widget.NewCheck("Use Basic Auth", func(checked bool) {
		if checked {
			setAuthenticationType("Basic")
		} else {
			setAuthenticationType("")
		}
	})
	tokenEntry := widget.NewEntry()
	tokenEntry.SetPlaceHolder("Enter Token")
	bearerCheckbox.Disable()
	basicAuthCheckbox.Disable()
	tokenEntry.Disable()

	settingAuthCheck := widget.NewCheck("Set Authentication", func(checked bool) {
		if checked {
			bearerCheckbox.Enable()
			basicAuthCheckbox.Enable()
			tokenEntry.Enable()
		} else {
			setAuthenticationType("")
			setAuthenticationToken("")
			bearerCheckbox.Disable()
			basicAuthCheckbox.Disable()
			tokenEntry.Disable()
			uiFunctions.NotificationPopUp("Authentication Cleared", "Authentication has been cleared.", deps)
		}
	})
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Setting Auth?", Widget: container.NewVBox(settingAuthCheck)},
			{Text: "Authentication Type", Widget: container.NewVBox(bearerCheckbox, basicAuthCheckbox)},
			{Text: "Token", Widget: tokenEntry},
		},
	}
	dialogContent := container.NewVBox(form)
	dialogContent.Resize(fyne.NewSize(400, 250))
	d := dialog.NewCustomConfirm("Set Authentication", "Save", "Cancel", dialogContent, func(b bool) {
		if b {
			if settingAuthCheck.Checked {
				if bearerCheckbox.Checked && basicAuthCheckbox.Checked {
					uiFunctions.NotificationPopUp("Error", "Please select only one authentication type.", deps)
					return
				}
				if !(tokenEntry.Text == "") {
					setAuthenticationToken(tokenEntry.Text)
					authSet = true
					return
				} else {
					uiFunctions.NotificationPopUp("Error", "Please enter a token.", deps)
					return
				}
			}
			if !settingAuthCheck.Checked {
				setAuthenticationType("")
				setAuthenticationToken("")
				uiFunctions.NotificationPopUp("Authentication Cleared", "Authentication has been cleared.", deps)
			}
		}
	}, deps.MainWindow)
	d.Resize(fyne.NewSize(400, 250))
	d.Show()
}

func SaveRequest(name, url, requestType string, headers map[string]string, body, authType, authToken string, deps shared.AppDependencies) {
	if !strings.Contains(name, "GET ") && !strings.Contains(name, "PUT ") && !strings.Contains(name, "POST ") && !strings.Contains(name, "PATCH ") && !strings.Contains(name, "DELETE ") {
		name = requestType + " " + name
	}
	savedRequest := SavedRequest{
		Name:                name,
		URL:                 url,
		RequestType:         requestType,
		Headers:             headers,
		Body:                body,
		AuthenticationType:  authType,
		AuthenticationToken: authToken,
	}
	savedRequests := deps.App.Preferences().String(savedRequestsPreferenceKey)
	var requests []SavedRequest
	if savedRequests != "" {
		err := json.Unmarshal([]byte(savedRequests), &requests)
		if err != nil {
			uiFunctions.NotificationPopUp("Error", "Failed to load saved requests: "+err.Error(), deps)
			return
		}
	}
	requests = append(requests, savedRequest)
	requestsJSON, err := json.Marshal(requests)
	if err != nil {
		uiFunctions.NotificationPopUp("Error", "Failed to save request. "+err.Error(), deps)
		return
	}
	deps.App.Preferences().SetString(savedRequestsPreferenceKey, string(requestsJSON))
	uiFunctions.NotificationPopUp("Request Saved", "Requests has been saved successfully.", deps)
}
