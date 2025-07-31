package configureRunner

import (
	"sync"
	"testAutomationSuiteGO/tests/apiTesting/apiTesting"
	"testAutomationSuiteGO/tests/e2e/fullScenarioTests"
	"testAutomationSuiteGO/tests/e2e/uiTesting"
	"testAutomationSuiteGO/tests/mobile/mobileTests"
)

func TestSectorRunConfig(testArea string) {

	if testArea == "All" {
		apiTesting.ExecuteAPITestsSequentially()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			uiTesting.ExecuteUITests()
			mobileTests.ExecuteMobileTests()
			fullScenarioTests.ExecuteFullScenarioTestsSequentially()
		}()
		wg.Wait()

		return
	}
	if testArea == "API" {
		apiTesting.ExecuteAPITestsSequentially()
		return
	}
	if testArea == "UI" {
		uiTesting.ExecuteUITestsSequentially()
		return
	}
	if testArea == "Mobile" {
		mobileTests.ExecuteMobileTestsSequentially()
		return
	}
	if testArea == "Full_Scenario" {
		fullScenarioTests.ExecuteFullScenarioTestsSequentially()
		return
	}
}
