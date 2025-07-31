package main

import (
	"database/sql"
	"fmt"
	"log"
	"testAutomationSuiteGO/internal/testingToolkit"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dbPath := testingToolkit.CurrPath() + "/app/testWriterUI/tests.db"
	tableName := "tests"
	columnName := "testType"

	AddColumnToTable(tableName, columnName, dbPath)
}

func AddColumnToTable(tableName, columnName, dbPath string) {

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	alterTableSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s TEXT;", tableName, columnName)
	_, err = db.Exec(alterTableSQL)
	if err != nil {
		log.Fatalf("Error adding column: %v", err)
	}

	fmt.Printf("Column '%s' added successfully\n", columnName)
}

func CreateTestCaseTable() {
	query := `CREATE TABLE test_cases (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tc_number VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    expected_outcome TEXT,
    test_data TEXT,
    account_int INTEGER DEFAULT 0,
    failure_statement TEXT,
    test_sector VARCHAR(100),
    single_test_run_status BOOLEAN DEFAULT TRUE,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	println(query)
}
