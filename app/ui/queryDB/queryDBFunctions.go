package queryDB

import (
	"database/sql"
	"fmt"
	"log"
	"testAutomationSuiteGO/app/uiFunctions"
	"testAutomationSuiteGO/internal/testingToolkit"
)

func GetAllColumnsFromATable(tableName, connectionString string) []string {
	var columnNames []string

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer db.Close()

	rows, err := db.Query(`
        SELECT column_name, data_type 
        FROM information_schema.columns 
        WHERE table_schema = 'public' 
        AND table_name = $1`, tableName)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var columnName, dataType string
		if err := rows.Scan(&columnName, &dataType); err != nil {
			log.Fatal(err)
		}
		columnNames = append(columnNames, columnName)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return columnNames
}

func GetTablesFromConfigJson(dbConnectionString string) []string {
	var tableNames []string
	configFile := testingToolkit.CurrPath() + "/config/uiArgs/uiArgs.json"
	var uiArgsRoot uiFunctions.UIArgsRoot
	err := uiFunctions.ReadJSONFile(configFile, &uiArgsRoot)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}
	dbConnections := uiArgsRoot.UIArgs.DatabaseConnectionStrings
	for i := 0; i < len(dbConnections); i++ {
		if dbConnections[i].DBTables == nil {
			continue
		}
		if len(dbConnections[i].DBTables) == 0 {
			continue
		}
		dbConnection := dbConnections[i]
		if dbConnection.ConnectionString == dbConnectionString {
			for j := 0; j < len(dbConnection.DBTables); j++ {
				tableName := dbConnection.DBTables[j].TableName
				tableNames = append(tableNames, tableName)
			}
		}
	}
	return tableNames
}

func GetColumnsFromConfigJson(dbConnectionString, tableName string) []string {
	var columnNames []string
	configFile := testingToolkit.CurrPath() + "/config/uiArgs/uiArgs.json"
	var uiArgsRoot uiFunctions.UIArgsRoot
	err := uiFunctions.ReadJSONFile(configFile, &uiArgsRoot)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}
	dbConnections := uiArgsRoot.UIArgs.DatabaseConnectionStrings
	for i := 0; i < len(dbConnections); i++ {
		if dbConnections[i].DBTables == nil {
			continue
		}
		if len(dbConnections[i].DBTables) == 0 {
			continue
		}
		dbConnection := dbConnections[i]
		if dbConnection.ConnectionString == dbConnectionString {
			for j := 0; j < len(dbConnection.DBTables); j++ {
				if dbConnection.DBTables[j].TableName == tableName {
					for k := 0; k < len(dbConnection.DBTables[j].Columns); k++ {
						columnName := dbConnection.DBTables[j].Columns[k].ColumnName
						columnNames = append(columnNames, columnName)
					}
				}
			}
		}
	}
	return columnNames
}

func GetFirstTableFromFirstConnectionStringFromConfigJson() string {
	configFile := testingToolkit.CurrPath() + "/config/uiArgs/uiArgs.json"
	var uiArgsRoot uiFunctions.UIArgsRoot
	err := uiFunctions.ReadJSONFile(configFile, &uiArgsRoot)
	if err != nil {
		log.Println(err)
		fmt.Println(err)
	}
	dbConnections := uiArgsRoot.UIArgs.DatabaseConnectionStrings
	if len(dbConnections) == 0 {
		return ""
	}
	if len(dbConnections[0].DBTables) == 0 {
		return ""
	}
	return dbConnections[0].DBTables[0].TableName
}
