package queryDB

import (
	"strings"
	"testAutomationSuiteGO/app/shared"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var SelectConsoleOutput *CustomEntry

func SelectTab(deps shared.AppDependencies, selects *QueryDBSelects) fyne.CanvasObject {

	var existingSelections []string

	queryToSubmitLabel := widget.NewLabel("SELECT * FROM " + selects.TableSelect.Selected + ";")
	selects.AddTableChangeListener(func(selected string) {
		existingSelections = []string{}
		queryToSubmitLabel.SetText("SELECT * FROM " + selected + ";")
	})

	SelectConsoleOutput = NewCustomEntry(deps.MainWindow)
	SelectConsoleOutput.SetPlaceHolder("Select Console Output")

	consolePadded := container.NewPadded(SelectConsoleOutput)

	titleLabel := widget.NewLabel("Select Query Builder:")

	addRequestedDataLabel := widget.NewLabel("Add Requested Data: ")

	buttonSetRequestedData := widget.NewButton("Set Requested Data", func() {
		CreateSelectColumnSelectorDialog(deps.MainWindow, selects.TableSelect.Selected, selects.ColumnSelect.Options, existingSelections, func(selected []string) {
			existingSelections = selected
			SelectConsoleOutput.SetText(StringifyRequestData(selected))
			if StringifyRequestData(selected) == "" {
				queryToSubmitLabel.SetText("SELECT * FROM " + selects.TableSelect.Selected + ";")
				return
			}
			queryToSubmitLabel.SetText("SELECT " + StringifyRequestData(selected) + " FROM " + selects.TableSelect.Selected + ";")
		})
	})
	dataRequestHBox := container.NewHBox(addRequestedDataLabel, buttonSetRequestedData)

	whereClauseLabel := widget.NewLabel("Where Clause: ")

	addFirstWhereClauseButton := widget.NewButton("Add First Condition", func() {
		AddFirstWhereClauseFunction(deps.MainWindow, selects.TableSelect.Selected, selects.ColumnSelect.Options, func(statement string) {
			currentQueryString := queryToSubmitLabel.Text
			if strings.Contains(currentQueryString, "WHERE") {
				currentQueryString = strings.Split(currentQueryString, " WHERE ")[0]
			}
			if strings.Contains(currentQueryString, ";") {
				currentQueryString = strings.Split(currentQueryString, ";")[0]
			}
			queryToSubmitLabel.SetText(currentQueryString + " " + statement + ";")
		})
	})

	addFollowingWhereOptionsButton := widget.NewButton("Add Following Conditions", func() {
	})

	addWhereClauseHBox := container.NewHBox(addFirstWhereClauseButton, addFollowingWhereOptionsButton)

	queryLabel := widget.NewLabel("Query:")

	queryToSubmitHBox := container.NewHBox(queryLabel, queryToSubmitLabel)

	submitButton := widget.NewButton("   Submit   ", func() {
	})

	resetQueryButton := widget.NewButton("Reset Query", func() {
		existingSelections = []string{}
		SelectConsoleOutput.SetText("")
		queryToSubmitLabel.SetText("SELECT * FROM " + selects.TableSelect.Selected + ";")
	})

	submitButtonHBox := container.NewHBox(submitButton, resetQueryButton)

	controls := container.NewVBox(titleLabel, dataRequestHBox, whereClauseLabel, addWhereClauseHBox, layout.NewSpacer(), queryToSubmitHBox, submitButtonHBox)

	split := container.NewVSplit(container.NewPadded(controls), consolePadded)

	contentWithConsole := container.NewBorder(nil, nil, nil, nil, split)

	return contentWithConsole
}

func StringifyRequestData(selectedColumns []string) string {
	var selectedColumnsString string
	for i, col := range selectedColumns {
		if i == 0 {
			selectedColumnsString = col
		} else {
			selectedColumnsString = selectedColumnsString + ", " + col
		}
	}
	return selectedColumnsString
}
