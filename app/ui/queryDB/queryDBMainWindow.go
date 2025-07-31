package queryDB

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type QueryDBSelects struct {
	DBSelect      *widget.Select
	TableSelect   *widget.Select
	ColumnSelect  *widget.Select
	tableHandlers []func(string)
}

type QueryDBWindowTab struct {
	name    string
	content fyne.CanvasObject
}

type CustomEntry struct {
	widget.Entry
	window fyne.Window
}

func NewCustomEntry(window fyne.Window) *CustomEntry {
	entry := &CustomEntry{window: window}
	entry.ExtendBaseWidget(entry)
	return entry
}

func AppendCustomConsoleForJsonPrettierPrint(consoleOutput *CustomEntry, text string) {
	currentText := consoleOutput.Text
	prettyText := uiFunctions.PrettyPrintJSON(text)
	consoleOutput.SetText(fmt.Sprintf("%s\n%s", currentText, prettyText))
}

func AppendToCustomConsole(consoleOutput *CustomEntry, text string) {
	currentText := consoleOutput.Text
	consoleOutput.SetText(fmt.Sprintf("%s\n%s", currentText, text))
}

func UpdateCustomConsole(consoleOutput *CustomEntry, text string) {
	consoleOutput.SetText(text)
}

func (e *CustomEntry) TappedSecondary(pe *fyne.PointEvent) {

	defaultMenuItems := []*fyne.MenuItem{
		fyne.NewMenuItem("Cut", func() {
			e.TypedShortcut(&fyne.ShortcutCut{Clipboard: e.window.Clipboard()})
		}),
		fyne.NewMenuItem("Copy", func() {
			e.TypedShortcut(&fyne.ShortcutCopy{Clipboard: e.window.Clipboard()})
		}),
	}

	contextMenu := fyne.NewMenu("", defaultMenuItems...)
	menu := widget.NewPopUpMenu(contextMenu, e.window.Canvas())
	menu.ShowAtPosition(pe.AbsolutePosition)

}

func QueryDBWindow(deps shared.AppDependencies) fyne.CanvasObject {

	dbSelect, tableSelect, columnsSelect := createLinkedSelects()
	if dbSelect == nil || tableSelect == nil || columnsSelect == nil {
		return container.NewCenter(widget.NewLabel("Error initializing database controls"))
	}

	selects := &QueryDBSelects{
		DBSelect:     dbSelect,
		TableSelect:  tableSelect,
		ColumnSelect: columnsSelect,
	}

	dbSelectLabel := widget.NewLabel("Database: ")
	tableSelectLabel := widget.NewLabel("Target Table: ")
	topControls := container.NewHBox(
		dbSelectLabel,
		dbSelect,
		tableSelectLabel,
		tableSelect,
		columnsSelect,
	)
	dbSelect.SetSelected(dbSelect.Options[0])

	tabs := []QueryDBWindowTab{
		{
			name:    "SELECT",
			content: SelectTab(deps, selects),
		},
		{
			name:    "UPDATE",
			content: UpdateTab(deps),
		},
		{
			name:    "INSERT",
			content: InsertTab(deps),
		},
		{
			name:    "DELETE",
			content: DeleteTab(deps),
		},
		{
			name:    "Query Writer",
			content: QueryWriterTab(deps),
		},
	}

	tabItems := make([]*container.TabItem, len(tabs))
	for i, tab := range tabs {
		tabItems[i] = container.NewTabItem(tab.name, tab.content)
	}

	tabsContainer := container.NewAppTabs(tabItems...)

	return container.NewBorder(topControls, nil, nil, nil, tabsContainer)
}

func createLinkedSelects() (*widget.Select, *widget.Select, *widget.Select) {
	dbTablesMap := make(map[string][]string)
	tableColumnsMap := make(map[string][]string)

	configFile := testingToolkit.CurrPath() + "/config/uiArgs/uiArgs.json"
	var uiArgsRoot uiFunctions.UIArgsRoot
	err := uiFunctions.ReadJSONFile(configFile, &uiArgsRoot)
	if err != nil {
		log.Println(err)
		return nil, nil, nil
	}

	for _, dbConn := range uiArgsRoot.UIArgs.DatabaseConnectionStrings {
		if dbConn.DBTables == nil {
			continue
		}
		var tableNames []string

		for _, table := range dbConn.DBTables {
			tableNames = append(tableNames, table.TableName)
			compositeKey := fmt.Sprintf("%s:%s", dbConn.DatabaseName, table.TableName)

			var columnNames []string
			for _, col := range table.Columns {
				columnNames = append(columnNames, col.ColumnName)
			}
			tableColumnsMap[compositeKey] = columnNames
		}

		dbTablesMap[dbConn.DatabaseName] = tableNames
	}

	dbSelect := widget.NewSelect(getPrimaryOptions(dbTablesMap), nil)
	tableSelect := widget.NewSelect([]string{}, nil)
	columnSelect := widget.NewSelect([]string{}, nil)

	dbSelect.OnChanged = func(selectedDB string) {
		if selectedDB == "" {
			return
		}

		newTableOptions := dbTablesMap[selectedDB]
		if len(newTableOptions) == 0 {
			return
		}

		tableSelect.Options = newTableOptions
		tableSelect.Selected = ""
		tableSelect.Refresh()

		columnSelect.Options = []string{}
		columnSelect.Selected = ""
		columnSelect.Refresh()

		if len(newTableOptions) > 0 {
			tableSelect.SetSelected(newTableOptions[0])
		}
	}

	tableSelect.OnChanged = func(selectedTable string) {
		if selectedTable == "" || dbSelect.Selected == "" {
			return
		}

		compositeKey := fmt.Sprintf("%s:%s", dbSelect.Selected, selectedTable)
		newColumnOptions := tableColumnsMap[compositeKey]

		columnSelect.Options = newColumnOptions
		columnSelect.Selected = ""
		columnSelect.Refresh()

		if len(newColumnOptions) > 0 {
			columnSelect.SetSelected(newColumnOptions[0])
		}
	}

	columnSelect.OnChanged = func(selectedColumn string) {
		if selectedColumn == "" {
			return
		}
	}

	return dbSelect, tableSelect, columnSelect
}

func getPrimaryOptions(dbMap map[string][]string) []string {
	var options []string
	for dbName := range dbMap {
		options = append(options, dbName)
	}
	sort.Strings(options)
	return options
}

func (s *QueryDBSelects) AddTableChangeListener(handler func(string)) {

	s.tableHandlers = append(s.tableHandlers, handler)

	originalHandler := s.TableSelect.OnChanged

	s.TableSelect.OnChanged = func(selected string) {

		if originalHandler != nil {
			originalHandler(selected)
		}

		for _, h := range s.tableHandlers {
			h(selected)
		}
	}

	for i := 0; i < len(s.tableHandlers)-1; i++ {
		if fmt.Sprintf("%p", s.tableHandlers[i]) == fmt.Sprintf("%p", originalHandler) {
			s.tableHandlers = append(s.tableHandlers[:i], s.tableHandlers[i+1:]...)
			break
		}
	}

}

func CreateSelectColumnSelectorDialog(window fyne.Window, tableName string, columns []string, existingSelections []string, onSave func([]string)) {
	var d *dialog.CustomDialog
	checkGroup := container.NewVBox()
	selectedColumns := make(map[string]bool)
	checkboxes := make(map[string]*widget.Check)
	var checkAllCheckbox *widget.Check

	selectionMap := make(map[string]bool)
	for _, selected := range existingSelections {
		selectionMap[selected] = true
		selectedColumns[selected] = true
	}

	for _, col := range columns {
		colName := col
		check := widget.NewCheck(col, func(checked bool) {
			selectedColumns[colName] = checked
			allChecked := true
			for _, col := range columns {
				if !selectedColumns[col] {
					allChecked = false
					break
				}
			}
			if checkAllCheckbox != nil {
				checkAllCheckbox.SetChecked(allChecked)
			}
		})
		check.SetChecked(selectionMap[col])
		if selectionMap[col] {
			selectedColumns[col] = true
		}
		checkboxes[colName] = check
		checkGroup.Add(check)
	}

	checkAllCheckbox = widget.NewCheck("Select All", func(checked bool) {
		for _, col := range columns {
			selectedColumns[col] = checked
		}
		for _, checkbox := range checkboxes {
			checkbox.SetChecked(checked)
		}
	})

	allSelected := true
	for _, col := range columns {
		if !selectedColumns[col] {
			allSelected = false
			break
		}
	}
	checkAllCheckbox.SetChecked(allSelected)

	scrollContainer := container.NewVScroll(checkGroup)
	scrollContainer.SetMinSize(fyne.NewSize(300, 200))

	saveBtn := widget.NewButton("Save", func() {
		var selected []string
		for _, col := range columns {
			if selectedColumns[col] {
				selected = append(selected, col)
			}
		}
		onSave(selected)
		d.Hide()
	})

	buttons := container.NewHBox(
		layout.NewSpacer(),
		saveBtn,
		layout.NewSpacer(),
	)

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("From \"%s\" edit the columns you want to query:", tableName)),
		checkAllCheckbox,
		scrollContainer,
		buttons,
	)

	d = dialog.NewCustom("Edit Selected Columns", "Cancel", content, window)
	d.Resize(fyne.NewSize(400, 300))
	d.Show()
}

func AddFirstWhereClauseFunction(window fyne.Window, tableName string, columns []string, onSave func(string)) {
	var d *dialog.CustomDialog

	columnSelect := widget.NewSelect(columns, nil)
	columnSelect.SetSelected(columns[0])

	columnSelectLabel := widget.NewLabel("Select Column: ")

	columnSelectHBox := container.NewHBox(columnSelectLabel, columnSelect)

	selectOperatorLabel := widget.NewLabel("Select Operator: ")

	operatorSelect := widget.NewSelect([]string{"=", "!=", ">", "<", ">=", "<=", "LIKE", "IN", "BETWEEN", "IS NULL"}, nil)

	operatorSelect.SetSelected("=")

	operatorSelectHBox := container.NewHBox(selectOperatorLabel, operatorSelect)

	conditionEntryLabel := widget.NewLabel("Enter Condition: ")

	conditionEntry := widget.NewEntry()

	entryContainer := container.NewBorder(nil, nil, conditionEntryLabel, nil, conditionEntry)

	operatorSelect.OnChanged = func(selected string) {
		switch selected {
		case "IS NULL":
			conditionEntry.SetText("")
			conditionEntry.Disable()
		case "BETWEEN":
			conditionEntry.SetText("value1 AND value2")
			conditionEntry.Enable()
		default:
			conditionEntry.SetText("")
		}
	}

	saveBtn := widget.NewButton("Add", func() {
		if operatorSelect.Selected == "BETWEEN" && !strings.Contains(conditionEntry.Text, "AND") {
			dialog.ShowError(errors.New("BETWEEN operator requires two values separated by 'AND'"), window)
			return
		}

		firstWhereClause := "WHERE " + columnSelect.Selected + " " + operatorSelect.Selected + " '" + conditionEntry.Text + "'"
		onSave(firstWhereClause)
		d.Hide()
	})

	buttons := container.NewHBox(
		layout.NewSpacer(),
		saveBtn,
		layout.NewSpacer(),
	)

	content := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Add the first where clause From The \"%s\" Table:", tableName)),
		columnSelectHBox,
		operatorSelectHBox,
		entryContainer,
		buttons,
	)

	d = dialog.NewCustom("Add The First Where Clause", "Cancel", content, window)
	d.Resize(fyne.NewSize(400, 300))
	d.Show()
}
