package testCaseStructuring

import (
	"os"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/internal/testRunParameters"
	testingToolkit "testAutomationSuiteGO/internal/testingToolkit"
)

// Returns tcNumber, expected, actual, AWSKEY, AWSSECRET
func TestCaseType1(tcNumber string) (string, string, string, string, string) {
	tests := RegressionTestJson
	var key string
	var secret string
	runLocal := testRunParameters.GetRunLocal()
	if runLocal && shared.GetCurrentEnvironment() == "production" {
		key = testingToolkit.GetENVVariable(testingToolkit.LocalVariablePath()+"/"+testRunParameters.GetLocalVariableFile(), "AWSKEYPRD")
		secret = testingToolkit.GetENVVariable(testingToolkit.LocalVariablePath()+"/"+testRunParameters.GetLocalVariableFile(), "AWSSECRETPRD")
	} else {
		if !runLocal {
			key = os.Args[3]
			secret = os.Args[4]
		}
	}
	tcNumber, expected, actual, _ := testCaseBuilder(tests, tcNumber)
	return tcNumber, expected, actual, key, secret
}

func testCaseBuilder(tests any, tcNumber string) (string, string, string, string) {

	var accountData string
	var expected string
	var actual string
	testCases := tests.(RegressionTestsJson).TestCases
	for i := range testCases {
		if testCases[i].TcNumber == tcNumber {
			expected = testCases[i].ExpectedOutcome
			actual = testCases[i].FailureStatement
			RegressionTestJson.TestCases[i].DateExecuted = testingToolkit.CurrentDate()
			RegressionTestJson.TestCases[i].StartTime = testingToolkit.CurrentTimeForTimeDurationWithMS()
			accountData = testCases[i].AccountData
			break
		}
	}
	return tcNumber, expected, actual, accountData
}
