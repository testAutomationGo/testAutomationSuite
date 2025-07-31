package notionAutomation

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

func SendReportToNotion() string {
	return ""
}

func GetCurrentMonthPageID(auth string) string {
	now := time.Now()
	currentMonth := now.Format("01/2006")
	pagesResponse := GetNotionPages(MainNotionReportingPage, auth)
	var pagesObject interface{}
	err := json.Unmarshal([]byte(pagesResponse), &pagesObject)
	if err != nil {
		return err.Error()
	}
	pages := pagesObject.(map[string]interface{})["results"].([]interface{})
	for _, page := range pages {
		pageObject := page.(map[string]interface{})
		println(pageObject["id"].(string))
		childPage := pageObject["child_page"].(map[string]interface{})
		title := childPage["title"].(string)
		if title == currentMonth {
			return pageObject["id"].(string)
		}
	}
	return "No page ID found"
}

func GetNotionPages(mainTestingResultsPageID, auth string) string {
	url := "https://api.notion.com/v1/blocks/" + mainTestingResultsPageID + "/children"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+auth)
	req.Header.Set("Notion-Version", "2021-05-13")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err.Error()
	}
	return string(responseBody)
}
