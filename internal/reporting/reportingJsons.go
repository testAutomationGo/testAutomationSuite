package reporting

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"testAutomationSuiteGO/internal/logger"
	testCaseStructuring "testAutomationSuiteGO/internal/testCaseStructuring"
	testRunParameters "testAutomationSuiteGO/internal/testRunParameters"
	testingToolkit "testAutomationSuiteGO/internal/testingToolkit"
	"time"
)

var timeToCompleteTests float64

type FullTestsJson struct {
	Tests Tests `json:"tests"`
}

type CompressedTestsJson struct {
	Tests Tests `json:"tests"`
}

type Test struct {
	TcNumber            string            `json:"tcNumber"`
	Title               string            `json:"title"`
	Description         string            `json:"description"`
	Steps               map[string]string `json:"steps"`
	ExpectedOutcome     string            `json:"ExpectedOutcome"`
	TestData            string            `json:"TestData"`
	AccountData         string            `json:"accountData"`
	ActualOutput        string            `json:"ActualOutput"`
	FailureStatement    string            `json:"FailureStatement"`
	Result              string            `json:"Result"`
	DateExecuted        string            `json:"DateExecuted"`
	StartTime           string            `json:"StartTime"`
	EndTime             string            `json:"EndTime"`
	MillisecondsTaken   string            `json:"MillisecondsTaken"`
	TestSector          string            `json:"TestSector"`
	SingleTestRunStatus bool              `json:"SingleTestRunStatus"`
	Notes               string            `json:"Notes"`
}

type Tests struct {
	TestCases          []Test `json:"testCases"`
	NumberOfTests      string `json:"NumberOfTests"`
	NumberOfTestsRan   string `json:"NumberOfTestsRan"`
	TestsPassed        string `json:"TestsPassed"`
	TestsFailed        string `json:"TestsFailed"`
	ExecutionStartTime string `json:"ExecutionStartTime"`
	ExecutionEndTime   string `json:"ExecutionEndTime"`
	ExecutionDate      string `json:"ExecutionDate"`
	ENV                string `json:"ENV"`
}

type TestSummary struct {
	NumberOfTests      string `json:"TotalNumberOfTests"`
	NumberOfTestsRan   string `json:"NumberOfTestsRan"`
	TestsPassed        string `json:"TestsPassed"`
	TestsFailed        string `json:"TestsFailed"`
	ExecutionStartTime string `json:"ExecutionStartTime"`
	ExecutionEndTime   string `json:"ExecutionEndTime"`
	ExecutionDate      string `json:"ExecutionDate"`
	ENV                string `json:"ENV"`
}

type FinalRegressionTestResults struct {
	testCaseStructuring.RegressionTestsJson
	NumberOfTests      string `json:"NumberOfTests"`
	NumberOfTestsRan   string `json:"NumberOfTestsRan"`
	TestsPassed        string `json:"TestsPassed"`
	TestsFailed        string `json:"TestsFailed"`
	ExecutionStartTime string `json:"ExecutionStartTime"`
	ExecutionEndTime   string `json:"ExecutionEndTime"`
	ExecutionDate      string `json:"ExecutionDate"`
	ENV                string `json:"ENV"`
}

func CreateResultsJsons() (string, []string) {
	resultStatement, failedTests := SetFullJsonWithResults()
	CompressResultsJson()
	return resultStatement, failedTests
}

func CompressResultsJson() {
	content, err := os.ReadFile(testRunParameters.GetResultsFolderPath() + "/fullJsonWithResults.json")
	if err != nil {
		log.Fatal(err)
	}

	var tests Tests
	err = json.Unmarshal(content, &tests)
	if err != nil {
		log.Fatal(err)
	}

	var filtered []Test
	for _, test := range tests.TestCases {
		if test.Result != "" {
			filtered = append(filtered, test)
		}
	}

	tests.TestCases = filtered

	updatedContent, err := json.MarshalIndent(tests, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(testRunParameters.GetResultsFolderPath()+"/results.json", updatedContent, 0644)
	if err != nil {
		log.Fatal(err)
	}
	GenerateHTMLReporter(testRunParameters.GetResultsFolderPath() + "/results.json")
}

func SetFullJsonWithResults() (string, []string) {
	existingJsonData := testCaseStructuring.RegressionTestJson
	startTime := testRunParameters.GetExecutionStartTime()
	endTime := SetExecutionEndTime(existingJsonData)
	finalTestResults := FinalRegressionTestResults{
		RegressionTestsJson: existingJsonData,
		NumberOfTests:       GetNumberOfTests(existingJsonData),
		NumberOfTestsRan:    GetNumberOfTestsRan(existingJsonData),
		TestsPassed:         GetTestsPassed(existingJsonData),
		TestsFailed:         GetTestsFailed(existingJsonData),
		ExecutionStartTime:  startTime,
		ExecutionEndTime:    endTime,
		ExecutionDate:       GetExecutionDate(existingJsonData),
		ENV:                 GetENV(),
	}
	jsonData, err := json.MarshalIndent(finalTestResults, "", "    ")
	if err != nil {
		fmt.Println(err)
		logger.Log(err.Error(), "Error Marshalling Final Test Results")
		return "Failed to generate results.", nil
	}
	file, err := os.Create(testRunParameters.GetResultsFolderPath() + "/fullJsonWithResults.json")
	if err != nil {
		fmt.Println(err)
		logger.Log(err.Error(), "Error Creating Full Json With Results File")
		return "Failed To generate results.", nil
	}
	defer file.Close()

	file.Write(jsonData)
	env := GetENV()
	numberOfTestsRan := GetNumberOfTestsRan(existingJsonData)
	testsPassed := GetTestsPassed(existingJsonData)
	testsFailed := GetTestsFailed(existingJsonData)
	logger.Log(env, "Environment")
	logger.Log(numberOfTestsRan, "Number of Tests Ran")
	logger.Log(testsPassed, "Tests Passed")
	logger.Log(testsFailed, "Tests Failed")
	logger.Log(GetExecutionDate(existingJsonData), "Execution Date")
	logger.Log(endTime, "Execution End Time")
	seconds, err := CalculateTimeDifference(startTime, endTime)
	timeToCompleteTests = seconds
	if err != nil {
		logger.Log(err.Error(), "Error Calculating Time Difference")
	}
	logger.Log(fmt.Sprintf("%.2f", seconds), "Seconds To Complete Tests")
	failedTestsSlice := []string{}
	logger.Log("", "Failed Tests")
	for i := range finalTestResults.TestCases {
		test := finalTestResults.TestCases[i]
		if test.Result == "Failed" {
			failedTestsSlice = append(failedTestsSlice, test.TcNumber+": "+test.Title)
			logger.Log(test.Result+" Title: "+test.Title, test.TcNumber)
		}
	}
	if testRunParameters.GetRunLocal() {
		logger.Log("Local Testing Mode: Test Sector Not Selected.", "Json Results")
		return "Test Execution In: " + env + " at " + startTime + "\nEnded At: " + endTime + "\nTotal Tests: " + numberOfTestsRan + "\nTests Passed: " + testsPassed + "\nTests Failed: " + testsFailed + "\nTime For Completion: " + testingToolkit.ConvertFloat64ToString(seconds, 2) + " Seconds", failedTestsSlice
	}
	testSector := os.Args[2]
	testSector = strings.ReplaceAll(testSector, "_", " ")
	return "Test Execution In: " + env + " For Test Sector: " + testSector + " Tests at " + startTime + " \nEnded At: " + endTime + "\nTotal Tests: " + numberOfTestsRan + "\nTests Passed: " + testsPassed + "\nTests Failed: " + testsFailed + "\nTime For Completion: " + testingToolkit.ConvertFloat64ToString(seconds, 2) + " Seconds", failedTestsSlice
}

func GetNumberOfTests(existingJsonData testCaseStructuring.RegressionTestsJson) string {
	numTests := len(existingJsonData.TestCases)
	return testingToolkit.ConvertIntToString(numTests)
}
func GetNumberOfTestsRan(existingJsonData testCaseStructuring.RegressionTestsJson) string {
	var numberOfTestsRan int
	for i := range existingJsonData.TestCases {
		testCase := existingJsonData.TestCases[i]
		if testCase.DateExecuted != "" {
			numberOfTestsRan++
		}
	}
	return testingToolkit.ConvertIntToString(numberOfTestsRan)
}

func GetTestsPassed(existingJsonData testCaseStructuring.RegressionTestsJson) string {
	var numTestsPassed int
	for i := range existingJsonData.TestCases {
		testCase := existingJsonData.TestCases[i]
		if testCase.Result == "Passed" {
			numTestsPassed++
		}
	}
	return testingToolkit.ConvertIntToString(numTestsPassed)
}

func GetTestsFailed(existingJsonData testCaseStructuring.RegressionTestsJson) string {
	var numberOfTestsFailed int
	for i := range existingJsonData.TestCases {
		testCase := existingJsonData.TestCases[i]
		if testCase.Result == "Failed" {
			numberOfTestsFailed++
		}
	}
	return testingToolkit.ConvertIntToString(numberOfTestsFailed)
}

func SetExecutionEndTime(existingJsonData testCaseStructuring.RegressionTestsJson) string {
	executionEndtime := testingToolkit.ReadableCurrentTimeWithMS()
	return executionEndtime
}

func GetExecutionDate(existingJsonData testCaseStructuring.RegressionTestsJson) string {
	return testingToolkit.CurrentDate()
}

func GetENV() string {
	return testRunParameters.GetEnvUppers()
}

func CalculateTimeDifference(startTime, endTime string) (float64, error) {
	layout := "2006-01-02 15:04:05.000"

	t1, err := time.Parse(layout, startTime)
	if err != nil {
		return 0, err
	}
	t2, err := time.Parse(layout, endTime)
	if err != nil {
		return 0, err
	}

	diff := t2.Sub(t1).Seconds()
	return diff, nil
}

func GetTimeToCompleteTests() float64 {
	return timeToCompleteTests
}
