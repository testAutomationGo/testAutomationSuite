package osVariables

import (
	"bufio"
	"log"
	"os"
	"strings"
	"testAutomationSuiteGO/internal/testingToolkit"
)

func SetLocalOSEnvVariables(env string) {
	if strings.Contains(env, "1") {
		if !testingToolkit.VerifyFileIsPresent(testingToolkit.CurrPath() + "/.env") {
			log.Println("This is a local environment and requires a .env file and is not found at this time.")
			os.Exit(0)
		}
		err := LoadEnv(testingToolkit.CurrPath() + "/.env")
		if err != nil {
			log.Println("Failed to load .env file:", err)
			os.Exit(0)
		}
	}
}

func LoadEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
	}
	return scanner.Err()
}
