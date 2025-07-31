package actionsUI

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testAutomationSuiteGO/internal/testingToolkit"
)

func GetTestSectorsOptions() []string {
	file, err := os.Open(testingToolkit.CurrPath() + "/.github/workflows/run-tests-V3.yml")
	if err != nil {
		return nil
	}
	defer file.Close()

	var options []string
	state := "searching" // searching -> found_test_area -> found_options -> reading_options

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch state {
		case "searching":
			if strings.Contains(trimmed, "test_area:") {
				state = "found_test_area"
			}
		case "found_test_area":
			if strings.Contains(trimmed, "options:") {
				state = "reading_options"
			}
		case "reading_options":
			if strings.Contains(trimmed, "jobs:") {
				return options
			}
			if strings.HasPrefix(trimmed, "- ") {
				option := strings.TrimPrefix(trimmed, "- ")
				options = append(options, option)
			} else if trimmed != "" {
				break
			}
		}
	}

	return options
}

func GetGitHubActionsToken() string {
	actionsFilePath := testingToolkit.CurrPath() + "/testData/accounts/githubActions.txt"
	file, err := os.Open(actionsFilePath)
	if err != nil {
		log.Println("Error opening file: ", err)
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ACTIONS=") {
			return strings.TrimPrefix(line, "ACTIONS=")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading file: ", err)
		return ""
	}

	return ""
}

func RunActions(token, owner, repo, workflowFile string, inputs map[string]string) (int, string) {
	payload := WorkflowDispatchRequest{
		Ref:    "main",
		Inputs: inputs,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Sprintf("error marshalling JSON: %v", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows/%s/dispatches",
		owner, repo, workflowFile)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return 0, fmt.Sprintf("error creating request: %v", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Sprintf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Sprintf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return resp.StatusCode, fmt.Sprintf("error: received status code %d, response: %s", resp.StatusCode, string(body))
	}

	fmt.Println("Workflow triggered successfully!")
	return resp.StatusCode, string(body)
}

type WorkflowDispatchRequest struct {
	Ref    string            `json:"ref"`
	Inputs map[string]string `json:"inputs"`
}
