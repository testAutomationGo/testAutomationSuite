package testWriterUI

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/ui/runParametersForUI"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/testWriter"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	_ "github.com/mattn/go-sqlite3"
)

type Test struct {
	Name     string
	Actions  []testWriter.Action
	TestType string
}

type TestWriterTab struct {
	name    string
	content fyne.CanvasObject
}

var WrittenTestsSelect *widget.Select

var testDB *sql.DB

var dbMutex sync.Mutex

func TestWriterConsoleRecordTestInstructions() string {
	return "Select an environment, enter a test name, and then click the \"Record Test\" button to start recording actions. When complete, click the \"Stop Recording\" button, or type \"ctrl + C\""
}

func CheckOnAndCreateDBAndTable() *sql.DB {

	dbFile := "testWriterUI/tests.db"

	isNewDB := false
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		file, err := os.Create(dbFile)
		if err != nil {
			uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not create database file: "+err.Error())
			log.Println("Could not create database file:", err)
			return nil
		}
		file.Close()
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "New \"tests.db\" was created.")
		isNewDB = true
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()
	testDB, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not open database: "+err.Error())
		log.Println("Could not open database:", err)
		return nil
	}

	createTable := `CREATE TABLE IF NOT EXISTS tests (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		actions TEXT NOT NULL,
		testType TEXT NOT NULL
	);`
	_, err = testDB.Exec(createTable)
	if err != nil {
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not create table: "+err.Error())
		log.Println("Could not create table:", err)
		testDB.Close()
		return nil
	}

	if isNewDB {
		sampleActions := []testWriter.Action{
			{Type: "info", Selector: "1", Value: "This is a sample test"},
			{Type: "info", Selector: "2", Value: "Create tests to begin."},
			{Type: "info", Selector: "3", Value: "This test will not run."},
		}
		actionsJSON, err := json.Marshal(sampleActions)
		if err != nil {
			uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not marshal sample actions: "+err.Error())
			log.Println("Could not marshal sample actions:", err)
			testDB.Close()
			return nil
		}

		_, err = testDB.Exec("INSERT INTO tests (name, actions, testType) VALUES (?, ?, ?)", "Example", actionsJSON, runParametersForUI.GetTestSectorNames()[0])
		if err != nil {
			uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not insert sample test: "+err.Error())
			log.Println("Could not insert sample test:", err)
			testDB.Close()
			return nil
		}
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Sample test data inserted.")
		log.Println("Sample test data inserted.")
	}

	uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Test DB is ready.")
	log.Println("Test DB is ready.")

	return testDB
}

func RecordTest(envName, testName string, testType string) {
	selectedEnvInt := shared.GetCurrentEnvInt()
	runParametersForUI.SetGUIEnvConfigurationsForAppFunctions(runParametersForUI.GetEnvironments()[selectedEnvInt])
	actions := testWriter.RecordActionsInTheBrowser(runParametersForUI.GetWebAppUrl(selectedEnvInt))
	WriteActionsToDB(testName, actions, testType)
	RefreshWrittenTestsSelect(testDB)
}

func WriteActionsToDB(testName string, actions []testWriter.Action, testType string) {
	actionsJson, err := json.Marshal(actions)
	if err != nil {
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not marshal actions: "+err.Error())
		log.Println("Could not marshal actions:", err)
		return
	}

	_, err = testDB.Exec("INSERT INTO tests (name, actions, testType) VALUES (?, ?, ?)", testName, string(actionsJson), string(testType))
	if err != nil {
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not insert actions into database: "+err.Error())
		log.Println("Could not insert actions into database:", err)
		return
	}

	uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, testName+": actions have been saved to database.")
	log.Println(testName + ": actions have been saved to database.")
}

func GetTestFromDB(testName string) (Test, error) {
	var actionsJson string
	var testType string
	row := testDB.QueryRow("SELECT actions, testType FROM tests WHERE name = ?", testName)
	if err := row.Scan(&actionsJson, &testType); err != nil {
		log.Println("Could not get test: " + err.Error())
		return Test{}, err
	}
	var actions []testWriter.Action
	if err := json.Unmarshal([]byte(actionsJson), &actions); err != nil {
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not unmarshal actions: "+err.Error())
		log.Println("Could not unmarshal actions:", err)
		return Test{}, err
	}
	return Test{Name: testName, Actions: actions, TestType: testType}, nil
}

func GetTestNames(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT name FROM tests")
	if err != nil {
		log.Println("Could not get test names: " + err.Error())
		println("Could not get test names: " + err.Error())
		return nil, err
	}
	defer rows.Close()

	var testNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Println("Could not scan test name: " + err.Error())
			return nil, err
		}
		testNames = append(testNames, name)
	}
	if err := rows.Err(); err != nil {
		log.Println("Could not get test names: " + err.Error())
		return nil, err
	}
	return testNames, nil
}

func GetActionsFromDB(testName string, db *sql.DB) ([]testWriter.Action, error) {
	var actionsJson string
	row := db.QueryRow("SELECT actions FROM tests WHERE name = ?", testName)
	if err := row.Scan(&actionsJson); err != nil {
		log.Println("Could not get actions: " + err.Error())
		return nil, err
	}

	var actions []testWriter.Action
	if err := json.Unmarshal([]byte(actionsJson), &actions); err != nil {
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not unmarshal actions: "+err.Error())
		log.Println("Could not unmarshal actions:", err)
		return nil, err
	}

	return actions, nil
}

func RefreshWrittenTestsSelect(db *sql.DB) {
	testNames, err := GetTestNames(db)
	if err != nil {
		log.Println("Could not get test names: " + err.Error())
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not get test names: "+err.Error())
	}
	WrittenTestsSelect.Options = testNames
	WrittenTestsSelect.Refresh()
	WrittenTestsSelect.SetSelectedIndex(0)
}

func ReplayTest(envName, testName string) {
	envOptions := runParametersForUI.GetEnvOptions()
	selectedEnvInt := uiFunctions.GetSelectIndex(envOptions, envName)
	runParametersForUI.SetGUIEnvConfigurationsForAppFunctions(runParametersForUI.GetEnvironments()[selectedEnvInt])
	actions, err := GetActionsFromDB(testName, testDB)
	for _, action := range actions {
		println(action.Type, action.Selector, action.Value)
	}
	if err != nil {
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not get actions: "+err.Error())
		log.Println("Could not get actions:", err)
		return
	}
	println(runParametersForUI.GetWebAppUrl(selectedEnvInt))
	err = testWriter.ReplayActions(actions, runParametersForUI.GetWebAppUrl(selectedEnvInt))
	if err != nil {
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Error replaying actions: "+err.Error())
		log.Println("Error replaying actions:", err)
	}
}

func RemoveTest(testName string) {
	_, err := testDB.Exec("DELETE FROM tests WHERE name = ?", testName)
	if err != nil {
		uiFunctions.AppendToGUIConsole(TestWriterConsoleOutput, "Could not remove test: "+err.Error())
		log.Println("Could not remove test:", err)
	}
	RefreshWrittenTestsSelect(testDB)
}

func EditTestView(testName string) {

	TestWriterConsoleOutput.Hide()

	consoleOutput := widget.NewMultiLineEntry()
	consoleOutput.SetPlaceHolder("Console Output")

	appendToConsole := func(text string) {
		currentText := consoleOutput.Text
		consoleOutput.SetText(fmt.Sprintf("%s\n%s", currentText, text))
	}

	var actions []testWriter.Action
	var testType string
	test, err := GetTestFromDB(testName)
	if err != nil {
		appendToConsole("Could not get test: " + err.Error())
		log.Println("Could not get test:", err)
		return
	}
	actions = test.Actions
	testType = test.TestType

	testNameLabel := widget.NewLabel("Test Name:")
	testNameEntry := widget.NewEntry()
	testNameEntry.SetText(testName)

	testTypeLabel := widget.NewLabel("Test Type:")
	testTypes := runParametersForUI.GetTestSectorNames()
	testTypeSelect := widget.NewSelect(testTypes, func(selected string) {})
	testTypeSelect.SetSelected(testType)

	testNameGridLayout := container.NewGridWithColumns(5, testNameLabel, testNameEntry, layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), testTypeLabel, testTypeSelect)

	indexLabelTitle := widget.NewLabel("Index Number:")
	typeLabelTitle := widget.NewLabel("Type of Actions:")
	selectorLabelTitle := widget.NewLabel("Selectors Used:")
	valueLabelTitle := widget.NewLabel("Values Used:")

	actionTable := container.New(layout.NewGridLayoutWithColumns(5))

	actionTable.Add(indexLabelTitle)
	actionTable.Add(typeLabelTitle)
	actionTable.Add(selectorLabelTitle)
	actionTable.Add(valueLabelTitle)
	actionTable.Add(layout.NewSpacer())

	for i, action := range actions {
		indexLabel := widget.NewLabel(fmt.Sprintf("%d.", i+1))
		typeActionOptions := []string{"click", "input", "wait"}
		typeSelect := widget.NewSelect(typeActionOptions, func(selected string) {})
		typeSelect.SetSelected(action.Type)
		selectorEntry := widget.NewEntry()
		selectorEntry.SetText(action.Selector)
		valueEntry := widget.NewEntry()
		valueEntry.SetPlaceHolder("If none, leave blank.")
		valueEntry.SetText(action.Value)

		actionTable.Add(indexLabel)
		actionTable.Add(typeSelect)
		actionTable.Add(selectorEntry)
		actionTable.Add(valueEntry)
		actionTable.Add(widget.NewButton("Remove", func() {
		}))
	}

	scrollableActionTable := container.NewScroll(actionTable)
	scrollableActionTable.SetMinSize(fyne.NewSize(600, 200))

	saveEditsButton := widget.NewButton("Save Edits", func() {
		SaveTestEdits()
		appendToConsole("Test edits saved.")
	})

	saveEditsButtonGridLayout := container.NewGridWithColumns(3, layout.NewSpacer(), saveEditsButton, layout.NewSpacer())
	topContainer := container.NewVBox(
		testNameGridLayout,
		widget.NewLabel("        "),
	)

	centerContainer := container.NewVBox(
		scrollableActionTable,
	)

	bottomContainer := container.NewVBox(
		layout.NewSpacer(),
		saveEditsButtonGridLayout,
		consoleOutput,
	)

	editorContent := container.NewGridWithRows(3, topContainer, centerContainer, bottomContainer)

	finalContent := container.NewBorder(nil, nil, nil, nil, editorContent)

	rightContent.Objects = []fyne.CanvasObject{finalContent}
	rightContent.Refresh()
}

func SaveTestEdits() {

}
