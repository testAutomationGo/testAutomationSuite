package logger

import (
	"testAutomationSuiteGO/internal/testRunParameters"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fmt"
	"log"
	"os"
)

func GenerateLogFile() {
	logFile := testRunParameters.GetResultsFolderPath() + "/logs.txt"
	file, err := os.Create(logFile)
	if err != nil {
		log.Fatal("Error creating log file: ", err)
	}
	defer file.Close()
}

func Log(logStatement string, tcNumber string) {
	//Log lines are trimmed to 500 characters to prevent logs and console logging from becoming too large
	//Log file will hold the entire log statement.
	//if not using tcNumber, this logger provides a ": " separator automatically, so it is not necessary for "Logging Time: " the ": " separator is not necesary
	var logStatementShortened string
	if len(logStatement) > 500 {
		logStatementShortened = logStatement[:500]
	}
	println(tcNumber + ": " + logStatementShortened)
	logFile := testRunParameters.GetResultsFolderPath() + "/logs.txt"
	present := testingToolkit.VerifyFileIsPresent(logFile)
	if !present {
		println("Log file not present")
		return
	}
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(tcNumber+"--Error opening log file at Log Function: ", err)
	}
	defer file.Close()
	if _, err := file.WriteString(tcNumber + ": " + logStatement + "\n"); err != nil {
		fmt.Println(tcNumber+"--Error writing to log file at Log Function: ", err)
	}
}
