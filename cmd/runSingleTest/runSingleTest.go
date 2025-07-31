package main

import (
	"os"
	"strings"
	"testAutomationSuiteGO/internal/logger"
	reporting "testAutomationSuiteGO/internal/reporting"
	"testAutomationSuiteGO/internal/slackAutomation"
	testCaseStructuring "testAutomationSuiteGO/internal/testCaseStructuring"
	testRunParameters "testAutomationSuiteGO/internal/testRunParameters"
	apiTesting "testAutomationSuiteGO/tests/apiTesting/apiTesting"
	"testAutomationSuiteGO/tests/e2e/fullScenarioTests"
	"testAutomationSuiteGO/tests/e2e/uiTesting"
	"testAutomationSuiteGO/tests/mobile/mobileTests"
)

func main() {

	testToRun := os.Args[2]

	testRunParameters.SetExecutionStartTime()
	testRunParameters.SetConfigFile(os.Args[1])
	testRunParameters.SetConfigRunnerJsonParameters()
	testRunParameters.SetResultsFolderName()
	testCaseStructuring.SetRegressionTestsJsonData()
	logger.GenerateLogFile()
	logger.Log(testToRun, "Running Single Test")
	logger.Log(testRunParameters.GetExecutionStartTime(), "Execution Start Time")

	RunTest(testToRun)
	//Report Results
	reportingStatement, failedTests := reporting.CreateResultsJsons()
	failedTestsMessage := slackAutomation.FailedMessagesFormatted(failedTests)
	if strings.Contains(os.Args[1], "1") {
		logger.Log("Local Testing Mode: Slack messages will not be sent.", "Slack Messaging")
		return
	}
	slackAutomation.SendMessageToSlack(reportingStatement+"\n"+failedTestsMessage, slackAutomation.GetSlackBotToken(), slackAutomation.GetSlackWebHookURL())

}

func RunTest(testToRun string) {

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
}

func ReturnSingleTestResult(tcNumber string) string {
	resultsJson := testCaseStructuring.GetRegressionTestsJsonData()
	testCases := resultsJson.TestCases
	var result string
	for i := range testCases {
		if testCases[i].TcNumber == tcNumber {
			result = testCases[i].Result
			break
		}
	}
	return result
}
