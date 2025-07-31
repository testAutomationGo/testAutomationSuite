package testsUI

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"
	"testAutomationSuiteGO/cmd/runRegressionAutomation/configureRunner"
	"testAutomationSuiteGO/internal/logger"
	"testAutomationSuiteGO/internal/reporting"
	"testAutomationSuiteGO/internal/slackAutomation"
	"testAutomationSuiteGO/internal/testCaseStructuring"
	"testAutomationSuiteGO/internal/testRunParameters"
	"testAutomationSuiteGO/internal/testingToolkit"
	"testAutomationSuiteGO/tests/apiTesting/apiTesting"
	"testAutomationSuiteGO/tests/e2e/fullScenarioTests"
	"testAutomationSuiteGO/tests/e2e/uiTesting"
	"testAutomationSuiteGO/tests/mobile/mobileTests"

	"fyne.io/fyne/v2/widget"
)

//If you create different test sectors or features, create the run functions here and ensure they are called with a button in runTestsWindow.go,
//include Single Test Run functions as well.

func RunFullRegression(envString string) string {
	SetRunParameters(envString)
	configureRunner.TestSectorRunConfig("All")
	reportingStatement, failedTests := reporting.CreateResultsJsons()
	failedTestsMessage := slackAutomation.FailedMessagesFormatted(failedTests)
	SetResultsFolderPath()
	return reportingStatement + "\n" + failedTestsMessage
}

func ExecuteTestSector(envString, testSector string) string {
	SetRunParameters(envString)
	configureRunner.TestSectorRunConfig(testSector)
	reportingStatement, failedTests := reporting.CreateResultsJsons()
	failedTestsMessage := slackAutomation.FailedMessagesFormatted(failedTests)
	SetResultsFolderPath()
	return reportingStatement + "\n" + failedTestsMessage
}

func RunSingleTest(envString, testToRun string) string {
	SetRunParameters(envString)
	switch {
	case strings.Contains(testToRun, "API"):
		apiTesting.ExecuteSingleAPITest(testToRun)
	case strings.Contains(testToRun, "UI"):
		uiTesting.ExecuteSingleUITest(testToRun)
	case strings.Contains(testToRun, "MB"):
		mobileTests.ExecuteSingleMobileTest(testToRun)
	case strings.Contains(testToRun, "FS"):
		fullScenarioTests.ExecuteSingleFullScenarioTest(testToRun)
	}
	reportingStatement, failedTests := reporting.CreateResultsJsons()
	failedTestsMessage := slackAutomation.FailedMessagesFormatted(failedTests)
	SetResultsFolderPath()
	return TestCaseLogs(testToRun) + "\n" + reportingStatement + "\n" + failedTestsMessage
}

func SetRunParameters(envString string) {
	testRunParameters.SetExecutionStartTime()
	testRunParameters.SetConfigFile(envString + "1")
	testRunParameters.SetConfigRunnerJsonParameters()
	testRunParameters.SetResultsFolderName()
	testCaseStructuring.SetRegressionTestsJsonData()
	logger.GenerateLogFile()
	logger.Log(testRunParameters.GetExecutionStartTime(), "Execution Start Time")
}

var ResultsFolderPath string

func SetResultsFolderPath() {
	ResultsFolderPath = testRunParameters.GetResultsFolderPath()
}

func GetResultsFolderPath() string {
	return ResultsFolderPath
}

func CheckSingleTestEntry(testToRun string) bool {
	if len(testToRun) > 10 {
		return false
	}
	if !strings.HasPrefix(testToRun, "TC_") {
		return false
	}
	return true
}

func UpdateSTWorkflow() string {
	var runningTestCases []string
	fileData, err := os.ReadFile(testingToolkit.CurrPath() + "/RegressionJSONTestCases.json")
	if err != nil {
		log.Fatal(err)
	}
	var data map[string]RootObject
	err = json.Unmarshal(fileData, &data)
	if err != nil {
		log.Fatal(err)
	}
	for _, obj := range data {
		for _, testCase := range obj.TestCases {
			if testCase.SingleTestRunStatus {
				runningTestCases = append(runningTestCases, "        - "+testCase.TcNumber)
			}
		}
	}
	ymlContent, err := os.Open(testingToolkit.CurrPath() + "/.github/workflows/run-single-test.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer ymlContent.Close()
	var lines []string
	scanner := bufio.NewScanner(ymlContent)
	occurrences := 0
	skipLines := false
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "options:") {
			occurrences++
			if occurrences == 2 {
				lines = append(lines, line)
				lines = append(lines, runningTestCases...)
				skipLines = true
				continue
			}
		}
		if strings.Contains(line, "jobs:") {
			skipLines = false
		}
		if !skipLines {
			lines = append(lines, line)
		}
	}
	outputContent := strings.Join(lines, "\n")
	err = os.WriteFile(testingToolkit.CurrPath()+"/.github/workflows/run-single-test.yml", []byte(outputContent), 0644)
	if err != nil {
		log.Fatal(err)
	}
	return "Success"
}

func ViewTestCaseListInConsole(outputConsole *widget.Entry) {
	testCaseListFile := testingToolkit.CurrPath() + "/testCaseSimpleList.txt"
	file, err := os.Open(testCaseListFile)
	if err != nil {
		log.Println("Error opening file: ", err)
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error scanning file: ", err)
		return
	}

	for _, line := range lines {
		if line == "" {
			break
		} else {
			outputConsole.Append(line + "\n")
		}
	}
}

type RootObject struct {
	TestCases []TestCase `json:"testCases"`
}

type TestCase struct {
	TcNumber            string            `json:"tcNumber"`
	Title               string            `json:"title"`
	Description         string            `json:"description"`
	Steps               map[string]string `json:"steps"`
	ExpectedOutcome     string            `json:"ExpectedOutcome"`
	TestData            string            `json:"TestData"`
	AccountInt          string            `json:"accountInt"`
	ActualOutput        string            `json:"ActualOutput"`
	FailureStatement    string            `json:"FailureStatement"`
	Result              string            `json:"Result"`
	DateExecuted        string            `json:"DateExecuted"`
	StartTime           string            `json:"StartTime"`
	EndTime             string            `json:"EndTime"`
	MillisecondsTaken   string            `json:"MillisecondsTaken"`
	SingleTestRunStatus bool              `json:"SingleTestRunStatus"`
	Notes               string            `json:"Notes"`
}

func TestCaseLogs(tcNumber string) string {
	filePath := testRunParameters.GetResultsFolderPath() + "/logs.txt"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var result strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, tcNumber) {
			if len(line) > 500 {
				line = line[:500]
			}
			result.WriteString(line + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return result.String()
}
