package main

import (
	"os"
	"strings"
	"testAutomationSuiteGO/cmd/runRegressionAutomation/configureRunner"
	"testAutomationSuiteGO/internal/logger"
	reporting "testAutomationSuiteGO/internal/reporting"
	"testAutomationSuiteGO/internal/slackAutomation"
	testCaseStructuring "testAutomationSuiteGO/internal/testCaseStructuring"
	testRunParameters "testAutomationSuiteGO/internal/testRunParameters"
)

func main() {
	//Set All Run Parameters
	testRunParameters.SetExecutionStartTime()
	testRunParameters.SetConfigFile(os.Args[1])
	testRunParameters.SetConfigRunnerJsonParameters()
	testRunParameters.SetResultsFolderName()
	testCaseStructuring.SetRegressionTestsJsonData()
	logger.GenerateLogFile()
	logger.Log(testRunParameters.GetExecutionStartTime(), "Execution Start Time")

	configureRunner.TestSectorRunConfig(os.Args[2])
	//Report Results
	reportingStatement, failedTests := reporting.CreateResultsJsons()
	failedTestsMessage := slackAutomation.FailedMessagesFormatted(failedTests)
	logger.Log(testRunParameters.GetExecutionStartTime(), reportingStatement)
	logger.Log(testRunParameters.GetExecutionStartTime(), failedTestsMessage)
	if strings.Contains(os.Args[1], "1") {
		logger.Log("Local Testing Mode: Slack messages will not be sent.", "Slack Messaging")
		return
	}
	slackAutomation.SendMessageToSlack(reportingStatement+"\n"+failedTestsMessage, slackAutomation.GetSlackBotToken(), slackAutomation.GetSlackWebHookURL())
}
