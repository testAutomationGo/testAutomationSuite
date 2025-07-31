package testsUI

import (
	"fmt"
	"os"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/ollamaInternal"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func RunFullTestSets(deps shared.AppDependencies) fyne.CanvasObject {

	consoleOutput := widget.NewMultiLineEntry()
	consoleOutput.SetPlaceHolder("Console Output:")

	appendToConsole := func(text string) {
		currentText := consoleOutput.Text
		consoleOutput.SetText(currentText + "\n" + text)
	}

	testCase := widget.NewEntry()
	testCase.SetPlaceHolder("EX: TC_API0001")

	column1 := container.NewVBox(
		widget.NewButton("Run Full Test Suite", func() {
			appendToConsole(shared.GetCurrentEnvironment() + " environment selected.")
			appendToConsole("Running Full Test Suite for " + shared.GetCurrentEnvironment() + " environment.")
			go func() {
				appendToConsole(RunFullRegression(shared.GetCurrentEnvironment()))
			}()
		}),
		widget.NewLabel("Run Single Test:"),
		testCase,
		widget.NewButton("Run Test", func() {
			appendToConsole("")
			if !CheckSingleTestEntry(testCase.Text) {
				appendToConsole("Test Case Entry is not valid. Please enter a valid Test Case Entry.")
				return
			}

			testCaseEntry := testCase.Text
			appendToConsole("Running Single Test: " + testCaseEntry + " for " + shared.GetCurrentEnvironment() + " environment.")
			go appendToConsole(RunSingleTest(shared.GetCurrentEnvironment(), testCaseEntry))
		}),
	)

	column2 := container.NewVBox(
		//widget.NewLabel(""), /* Spacer */
		widget.NewLabel(""), /* Spacer */
		// When adding new test sectors, ensure to add a button here to call the function.
		widget.NewLabel("Select Test Section:"),
		widget.NewButton("API", func() {
			appendToConsole("")
			appendToConsole("Running API Tests.")
			go appendToConsole(ExecuteTestSector(shared.GetCurrentEnvironment(), "API"))
			appendToConsole("API Tests have completed.")
		}),
		widget.NewButton("UI", func() {
			appendToConsole("")
			appendToConsole("Running UI Tests.")
			go appendToConsole(ExecuteTestSector(shared.GetCurrentEnvironment(), "UI"))
			appendToConsole("UI Tests have completed.")
		}),
		widget.NewButton("Mobile", func() {
			appendToConsole("")
			appendToConsole("Running Mobile Tests.")
			go appendToConsole(ExecuteTestSector(shared.GetCurrentEnvironment(), "Mobile"))
			appendToConsole("Mobile Tests have completed.")
		}),
		widget.NewButton("Full Scenario Tests", func() {
			appendToConsole("")
			appendToConsole("Running Full Scenario Tests.")
			go appendToConsole(ExecuteTestSector(shared.GetCurrentEnvironment(), "Full_Scenario"))
			appendToConsole("Full Scenario Tests have completed.")
		}),
	)

	bottomButtonViewTestCaseList := widget.NewButton("View Test Case List", func() {
		consoleOutput.SetText("Test Case List:\n")
		ViewTestCaseListInConsole(consoleOutput)
	})

	bottomButtonResultsFile := widget.NewButton("Open Results File", func() {
		go uiFunctions.OpenFile(GetResultsFolderPath() + "/results.json")
	})

	bottomButtonResultsFolder := widget.NewButton("Open Results Folder", func() {
		go uiFunctions.OpenFolder(GetResultsFolderPath())
	})

	bottomButtonGenerateExecutiveSummary := widget.NewButton("Generate Executive Summary", func() {
		go func() {
			appendToConsole("Generating Executive Summary...")
			htmlFile := GetResultsFolderPath() + "/test_results.html"

			if _, err := os.Stat(htmlFile); os.IsNotExist(err) {
				appendToConsole(fmt.Sprintf("Error: HTML file '%s' does not exist\n", htmlFile))
				return
			}

			analyzer := ollamaInternal.NewTestAnalyzer()

			err := analyzer.CheckOllamaInstalled()
			if err != nil {
				appendToConsole(fmt.Sprintf("Error: %v\n", err))
				return
			}

			err = analyzer.CheckModelExists()
			if err != nil {
				appendToConsole(fmt.Sprintf("Error: %v\n", err))
				return
			}

			err = analyzer.ProcessHTMLFile(htmlFile)
			if err != nil {
				appendToConsole(fmt.Sprintf("Error processing HTML file: %v\n", err))
				return
			}

			appendToConsole("Analysis complete!")
		}()
	})

	bottomButtonClearConsole := widget.NewButton("Clear Console", func() {
		consoleOutput.SetText("Console Output:")
	})

	buttonColumns := container.NewHBox(
		column1,
		column2,
	)

	leftContent := container.NewVBox(
		buttonColumns,
		layout.NewSpacer(),
		bottomButtonViewTestCaseList,
		bottomButtonResultsFile,
		bottomButtonResultsFolder,
		bottomButtonGenerateExecutiveSummary,
		bottomButtonClearConsole,
	)

	contentWithConsole := container.New(layout.NewBorderLayout(nil, nil, leftContent, nil), leftContent, consoleOutput)

	return contentWithConsole
}
