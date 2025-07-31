package slackAutomation

import (
	"os"
	apiFunctions "testAutomationSuiteGO/internal/apiFunctions"
	"testAutomationSuiteGO/internal/testRunParameters"
	"testAutomationSuiteGO/internal/testingToolkit"
)

func FailedMessagesFormatted(failedTests []string) string {
	var message string
	message += "Failed Tests:\n"
	for _, test := range failedTests {
		message += test + "\n"
	}
	if testRunParameters.GetRunLocal() {
		return message + "Local Testing Mode: Slack messages will not be sent."
	}
	message += "GitHub Actions Job: Go to <" + "https://github.com/PinataCloud/qa_automation/actions/runs/" + testRunParameters.GetRunID() + "|here>.\n"
	return message
}

func SendMessageToSlack(message, botToken, webhook string) {
	payload := "{\"text\": \"" + message + "\"}"
	apiFunctions.PostRequestWithJson(webhook, botToken, payload)
}

func GetSlackWebHookURL() string {
	if testRunParameters.GetRunLocal() {
		return testingToolkit.GetENVVariable(testingToolkit.CurrPath()+"/testData/localTestingVariables/slackReporting.txt", "SLACK_WEBHOOK_URL")
	}
	return os.Args[2]
}

func GetSlackBotToken() string {
	if testRunParameters.GetRunLocal() {
		return testingToolkit.GetENVVariable(testingToolkit.CurrPath()+"/testData/localTestingVariables/slackReporting.txt", "SLACK_BOT_TOKEN")
	}
	return os.Args[3]
}
