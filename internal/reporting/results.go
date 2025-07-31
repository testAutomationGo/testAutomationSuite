package reporting

import (
	"fmt"
	testCaseStructuring "testAutomationSuiteGO/internal/testCaseStructuring"
	testingToolkit "testAutomationSuiteGO/internal/testingToolkit"
	"time"
)

func AssertEquals(tcNumber, expectedStatement, actualStatement, actualOutput, testData string) {
	tests := testCaseStructuring.RegressionTestJson.TestCases

	for i, test := range tests {
		if test.TcNumber == tcNumber {
			if expectedStatement != actualStatement {
				tests[i].Result = "Failed"
			} else {
				tests[i].Result = "Passed"
			}
			tests[i].ActualOutput = actualOutput
			tests[i].EndTime = testingToolkit.CurrentTimeForTimeDurationWithMS()
			tests[i].MillisecondsTaken = GetTimeDifferenceMilliSeconds(tests[i].StartTime, tests[i].EndTime)
			if testData != "" {
				tests[i].TestData = testData
			} else {
				tests[i].TestData = ""
			}
			break
		}
	}
}

func GetTimeDifferenceMilliSeconds(executionStartTimeString, executionEndTimeString string) string {
	startTime, err := parseCustomFormat(executionStartTimeString)
	if err != nil {
		fmt.Println("Error parsing start time:", err.Error())
		return "Error parsing start time."
	}
	endTime, err := parseCustomFormat(executionEndTimeString)
	if err != nil {
		fmt.Println("Error parsing end time:", err.Error())
		return "Error parsing end time."
	}
	duration := endTime.Sub(startTime)
	return fmt.Sprintf("%d milliseconds", duration.Milliseconds())
}

func parseCustomFormat(timeString string) (time.Time, error) {
	layout := "20060102150405.000"
	parseTime, err := time.Parse(layout, timeString)
	if err != nil {
		fmt.Println("Error parsing time:", err.Error())
		return time.Time{}, err
	}
	return parseTime, nil
}
