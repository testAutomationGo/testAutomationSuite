package codeRunner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testAutomationSuiteGO/internal/testingToolkit"
	"time"
)

func RunGoCode(code, path string) (string, error) {
	mainPath := filepath.Join(testingToolkit.CurrPath(), path, "main.go")
	err := os.WriteFile(mainPath, []byte(code), 0644)
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(mainPath)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", mainPath)
	cmd.Dir = dir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command execution failed: %v, stderr: %s", err, errBuf.String())
	}
	return outBuf.String(), nil
}
