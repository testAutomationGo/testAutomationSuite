package logger

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

type LogData struct {
	Entries map[string][]string
	EnvData map[string]string
}

var logData LogData

func RetrieveLogDataFromLogFile(logFilePath string) {
	file, err := os.Open(logFilePath)
	if err != nil {
		log.Println("Error opening log file at RetrieveLogDataFromLogFile: ", err)
	}
	defer file.Close()

	logData = LogData{
		Entries: make(map[string][]string),
		EnvData: make(map[string]string),
	}

	scanner := bufio.NewScanner(file)
	var currentTC string
	envSection := false

	tcRegex := regexp.MustCompile(`^(TC_[A-Z0-9]+):\s*(.*)`)
	envRegex := regexp.MustCompile(`^Environment:`)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) > 500 {
			line = line[:500]
		}
		if envRegex.MatchString(line) {
			envSection = true
		}
		if envSection {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				logData.EnvData[key] = value
			}
			continue
		}

		matches := tcRegex.FindStringSubmatch(line)
		if len(matches) > 0 {
			currentTC = matches[1]
			logData.Entries[currentTC] = append(logData.Entries[currentTC], matches[2])
		} else if currentTC != "" {
			logData.Entries[currentTC] = append(logData.Entries[currentTC], line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading log file at RetrieveLogDataFromLogFile: ", err)
	}
}

func GetLogData() LogData {
	return logData
}
