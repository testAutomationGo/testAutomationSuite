package uiFunctions

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"testAutomationSuiteGO/app/shared"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type UIArgsRoot struct {
	UIArgs UIArgs `json:"uiArgs"`
}

type UIArgs struct {
	GmailCreds                string          `json:"gmailCreds"`
	DatabaseConnectionStrings []DBConnections `json:"databaseConnectionStrings"`
}

type DBConnections struct {
	DatabaseName     string     `json:"dbName"`
	ConnectionString string     `json:"connectionString"`
	DBTables         []DBTables `json:"dbTables"`
}

type DBTables struct {
	TableName string      `json:"dbTableName"`
	Columns   []DBColumns `json:"dbColumns"`
}

type DBColumns struct {
	ColumnName string `json:"columnName"`
	ColumnType string `json:"columnType"`
}

type Column struct {
	Name     string
	DataType string
}

type Table struct {
	Name    string
	Columns []Column
}

type DatabaseSchema struct {
	Tables map[string]*Table
}

func UpdateDBSchemas(myWindow fyne.Window, deps shared.AppDependencies) {
	// Create UI elements
	progressBar := widget.NewProgressBar()
	progressLabel := widget.NewLabel("Updating Database Schemas")
	content := container.NewVBox(
		progressLabel,
		progressBar,
	)

	d := dialog.NewCustom("Updating Database Schemas", "Close", content, myWindow)
	d.Resize(fyne.NewSize(300, 100))
	d.Show()

	go func() {
		updateSchemas(progressBar, progressLabel, myWindow, d)
	}()
}

func updateSchemas(progressBar *widget.ProgressBar, progressLabel *widget.Label,
	myWindow fyne.Window, d *dialog.CustomDialog) {

	configFile := testingToolkit.CurrPath() + "/config/uiArgs/uiArgs.json"
	var uiArgsRoot UIArgsRoot
	if err := ReadJSONFile(configFile, &uiArgsRoot); err != nil {
		log.Println("Error reading config file:", err)
		return
	}

	dbConnections := uiArgsRoot.UIArgs.DatabaseConnectionStrings
	totalDBs := float64(len(dbConnections))

	for i, dbConn := range dbConnections {
		updateProgressUI(progressBar, progressLabel, i, totalDBs, dbConn.DatabaseName, len(dbConnections))

		schema, err := fetchDatabaseSchema(dbConn.ConnectionString)
		if err != nil {
			log.Println("Error fetching schema:", err)
			continue
		}

		if err := updateConfiguration(configFile, i, schema); err != nil {
			log.Println("Error updating configuration:", err)
			continue
		}

		testingToolkit.DelaySeconds(3)
	}

	completeUpdate(progressBar, progressLabel, d, myWindow)
}

func fetchDatabaseSchema(connectionString string) (*DatabaseSchema, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	defer db.Close()

	query := `
        SELECT table_name, column_name, data_type 
        FROM information_schema.columns 
        WHERE table_schema = 'public' 
        AND table_name NOT LIKE 'pg_%' 
        ORDER BY table_name, ordinal_position;
    `

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	schema := &DatabaseSchema{
		Tables: make(map[string]*Table),
	}

	for rows.Next() {
		var tableName, columnName, dataType string
		if err := rows.Scan(&tableName, &columnName, &dataType); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		if _, ok := schema.Tables[tableName]; !ok {
			schema.Tables[tableName] = &Table{
				Name: tableName,
			}
		}
		schema.Tables[tableName].Columns = append(schema.Tables[tableName].Columns, Column{
			Name:     columnName,
			DataType: dataType,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return schema, nil
}

func updateConfiguration(configFile string, index int, schema *DatabaseSchema) error {
	config := &UIArgsRoot{}

	currentData, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	if err := json.Unmarshal(currentData, config); err != nil {
		return fmt.Errorf("error unmarshaling config: %v", err)
	}

	UpdateConfigWithSchema(config, index, schema)

	return SaveConfig(config, configFile)
}

func updateProgressUI(progressBar *widget.ProgressBar, progressLabel *widget.Label,
	current int, total float64, dbName string, totalCount int) {

	progress := float64(current) / total
	progressBar.SetValue(progress)
	progressLabel.SetText(fmt.Sprintf("Updating %s (%d/%d)",
		dbName, current+1, totalCount))
}

func completeUpdate(progressBar *widget.ProgressBar, progressLabel *widget.Label,
	d *dialog.CustomDialog, myWindow fyne.Window) {

	progressBar.SetValue(1)
	progressLabel.SetText("Update Complete!")
	d.Hide()
	dialog.ShowInformation("Database Schemas Updated",
		"Database Schemas have been updated", myWindow)
}

func UpdateConfigWithSchema(config *UIArgsRoot, dbIndex int, schema *DatabaseSchema) {
	if dbIndex >= len(config.UIArgs.DatabaseConnectionStrings) {
		return
	}

	var dbTables []DBTables
	for tableName, table := range schema.Tables {
		dbColumns := make([]DBColumns, len(table.Columns))
		for i, col := range table.Columns {
			dbColumns[i] = DBColumns{
				ColumnName: col.Name,
				ColumnType: col.DataType,
			}
		}

		dbTables = append(dbTables, DBTables{
			TableName: tableName,
			Columns:   dbColumns,
		})
	}

	config.UIArgs.DatabaseConnectionStrings[dbIndex].DBTables = dbTables
}

func SaveConfig(config *UIArgsRoot, filename string) error {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}
