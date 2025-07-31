package testCaseStructuring

import (
	"encoding/json"
	"fmt"
	"os"
	projectData "testAutomationSuiteGO/internal/projectData"
	testingToolkit "testAutomationSuiteGO/internal/testingToolkit"
)

var RegressionTestJson RegressionTestsJson

type RegressionTestsJson struct {
	TestCases          []TestCase `json:"testCases"`
	NumberOfTests      string     `json:"NumberOfTests"`
	NumberOfTestsRan   string     `json:"NumberOfTestsRan"`
	TestsPassed        string     `json:"TestsPassed"`
	TestsFailed        string     `json:"TestsFailed"`
	ExecutionStartTime string     `json:"ExecutionStartTime"`
	ExecutionEndTime   string     `json:"ExecutionEndTime"`
	ExecutionDate      string     `json:"ExecutionDate"`
	ENV                string     `json:"ENV"`
}

type TestCase struct {
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

func SetRegressionTestsJsonData() error {
	filePath := testingToolkit.CurrPath() + "/" + projectData.TestCaseJSONFileName
	jsonFile, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer jsonFile.Close()
	decoder := json.NewDecoder(jsonFile)
	err = decoder.Decode(&RegressionTestJson)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetRegressionTestsJsonData() RegressionTestsJson {
	return RegressionTestJson
}
