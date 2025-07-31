package reporting

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testAutomationSuiteGO/internal/testCaseStructuring"
	"testAutomationSuiteGO/internal/testData"
	"testAutomationSuiteGO/internal/testRunParameters"
	"testAutomationSuiteGO/internal/testingToolkit"
	"time"
)

const MainNotionReportingPage = "200e8162bffd49fbadfb6764bf5c06ff"

func ReportResultsToNotion(auth, resultsFolderFullPath string) string {
	reportingTitle := "Tested In " + testRunParameters.GetEnvString() + " on " + testingToolkit.CurrentTimeForNaming()
	currentMonthReportingPageID := GetCurrentMonthReportingPageID(auth)
	resultsBodyPrefix := ResultsRunPagePrefix(currentMonthReportingPageID)
	resultsFileJson := testData.ConvertJsonFileToObject(resultsFolderFullPath + "/results.json")
	var resultsJson testCaseStructuring.RegressionTestsJson
	err := json.Unmarshal(resultsFileJson, &resultsJson)
	if err != nil {
		return err.Error()
	}
	return GenerateNotionResultsSetPage(reportingTitle, resultsBodyPrefix, auth, resultsJson)
}

func GenerateNotionResultsSetPage(reportingTitle, resultsBodyPrefix, auth string, resultsJson testCaseStructuring.RegressionTestsJson) string {
	return "url"
}

func GetCurrentMonthReportingPageID(auth string) string {
	currentTime := time.Now().Format("01/2006")
	_, getPagesResponse := GetNotionPages(MainNotionReportingPage, auth)
	var currentPageReportingID string
	var notionPagesRoot NotionPagesRoot
	err := json.Unmarshal([]byte(getPagesResponse), &notionPagesRoot)
	if err != nil {
		return err.Error()
	}
	var currentMonthPageExists bool
	for _, page := range notionPagesRoot.Results {
		if page.ChildPage.Title == currentTime {
			currentPageReportingID = page.ID
			currentMonthPageExists = true
		}
	}
	if !currentMonthPageExists {
		currentPageReportingID, err = NotionMonthlyReportingPagePOST(MainNotionReportingPage, auth, currentTime)
		if err != nil {
			return err.Error()
		}
	}
	return currentPageReportingID
}

func NotionMonthlyReportingPagePOST(mainTestingResultsPageID, auth, currentMonth string) (string, error) {

	requestBody := PageRequest{}
	requestBody.Parent.Type = "page_id"
	requestBody.Parent.PageID = mainTestingResultsPageID
	requestBody.Properties.Title.ID = "title"
	requestBody.Properties.Title.Type = "title"
	requestBody.Properties.Title.Title = []struct {
		Type string `json:"type"`
		Text struct {
			Content string `json:"content"`
		} `json:"text"`
	}{
		{
			Type: "text",
			Text: struct {
				Content string `json:"content"`
			}{
				Content: currentMonth,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	url := "https://api.notion.com/v1/pages"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+auth)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response code: %d, body: %s", resp.StatusCode, string(body))
	}

	var responseObject map[string]interface{}
	err = json.Unmarshal(body, &responseObject)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response: %v", err)
	}

	pageID, ok := responseObject["id"].(string)
	if !ok {
		return "", fmt.Errorf("page ID not found in response")
	}
	return pageID, nil
}

func GetNotionPages(pageID, auth string) (int, string) {
	url := "https://api.notion.com/v1/blocks/" + pageID + "/children"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+auth)
	req.Header.Set("Notion-Version", "2021-05-13")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}
	return resp.StatusCode, string(responseBody)
}

func ResultsRunPagePrefix(currentMonthReportingPageID string) string {
	//This JSON part hold the parent section of payload.
	return "{\r\n" +
		"  \"parent\": {\r\n" +
		"    \"type\": \"page_id\",\r\n" +
		"    \"page_id\": \"" + currentMonthReportingPageID + "\"\r\n" +
		"  },\r\n"
}

func ToggleBlockTooLong(fileName, resultsFolderFullPath string) string {
	// Toggles only allow for 100 lines of text to be hidden in the toggle.
	fileLines := []string{}
	file, err := os.Open(resultsFolderFullPath + "/" + fileName)
	if err != nil {
		fmt.Println("--", err)
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileLines = append(fileLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("--", err)
		return ""
	}

	howManyNewLists := (len(fileLines) + 99) / 100
	sublists := make([][]string, howManyNewLists)

	for i := range sublists {
		start := i * 100
		end := start + 100
		if end > len(fileLines) {
			end = len(fileLines)
		}
		sublists[i] = fileLines[start:end]
	}

	var fileToggleBlock strings.Builder
	for i, sublist := range sublists {
		if len(sublists) == 1 {
			fileToggleBlock.WriteString(FileJson(fileName, sublist))
			break
		} else {
			fileToggleBlock.WriteString(FileJson(fmt.Sprintf("Part %d of %s", i+1, fileName), sublist))
		}
		if i != len(sublists)-1 {
			fileToggleBlock.WriteString(",\r\n")
		}
	}

	return fileToggleBlock.String()
}

func FileJson(fileName string, fileLines []string) string {
	tb := ToggleBlock{
		Object: "block",
		Type:   "toggle",
		Toggle: struct {
			RichText []struct {
				Type string `json:"type"`
				Text struct {
					Content string `json:"content"`
					Link    any    `json:"link"`
				} `json:"text"`
				Annotations struct {
					Underline bool   `json:"underline"`
					Bold      bool   `json:"bold"`
					Color     string `json:"color"`
				} `json:"annotations"`
			} `json:"rich_text"`
			Children []struct {
				Object    string `json:"object"`
				Type      string `json:"type"`
				Paragraph struct {
					RichText []struct {
						Type string `json:"type"`
						Text struct {
							Content string `json:"content"`
						} `json:"text"`
					} `json:"rich_text"`
				} `json:"paragraph"`
			} `json:"children"`
		}{
			RichText: []struct {
				Type string `json:"type"`
				Text struct {
					Content string `json:"content"`
					Link    any    `json:"link"`
				} `json:"text"`
				Annotations struct {
					Underline bool   `json:"underline"`
					Bold      bool   `json:"bold"`
					Color     string `json:"color"`
				} `json:"annotations"`
			}{
				{
					Type: "text",
					Text: struct {
						Content string `json:"content"`
						Link    any    `json:"link"`
					}{
						Content: fileName,
						Link:    nil,
					},
					Annotations: struct {
						Underline bool   `json:"underline"`
						Bold      bool   `json:"bold"`
						Color     string `json:"color"`
					}{
						Underline: true,
						Bold:      true,
						Color:     "default",
					},
				},
			},
			Children: []struct {
				Object    string `json:"object"`
				Type      string `json:"type"`
				Paragraph struct {
					RichText []struct {
						Type string `json:"type"`
						Text struct {
							Content string `json:"content"`
						} `json:"text"`
					} `json:"rich_text"`
				} `json:"paragraph"`
			}{
				{
					Object: "block",
					Type:   "paragraph",
					Paragraph: struct {
						RichText []struct {
							Type string `json:"type"`
							Text struct {
								Content string `json:"content"`
							} `json:"text"`
						} `json:"rich_text"`
					}{
						RichText: fileLinesBlock(fileLines),
					},
				},
			},
		},
	}

	data, err := json.Marshal(tb)
	if err != nil {
		return ""
	}

	return string(data)
}

func fileLinesBlock(fileLines []string) []struct {
	Type string `json:"type"`
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
} {
	var richText []struct {
		Type string `json:"type"`
		Text struct {
			Content string `json:"content"`
		} `json:"text"`
	}

	for _, line := range fileLines {
		if len(line) > 1900 {
			line = line[:1900]
		}
		line = strings.ReplaceAll(line, "\"", "\\\"")
		richText = append(richText, struct {
			Type string `json:"type"`
			Text struct {
				Content string `json:"content"`
			} `json:"text"`
		}{
			Type: "text",
			Text: struct {
				Content string `json:"content"`
			}{
				Content: "   " + line + "\n",
			},
		})
	}

	return richText
}

func NotionPageResultSetPOST(jsonPayload, auth string) (string, error) {
	fullResponse := ""

	url := "https://api.notion.com/v1/pages"

	var requestBody bytes.Buffer
	err := json.Compact(&requestBody, []byte(jsonPayload))
	if err != nil {
		return fullResponse, fmt.Errorf("failed to compact JSON payload: %v", err)
	}

	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return fullResponse, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+auth)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fullResponse, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fullResponse, fmt.Errorf("failed to read response body: %v", err)
	}

	fullResponse = fmt.Sprintf("Status Code: %d + Body: %s", resp.StatusCode, string(body))

	return fullResponse, nil
}
