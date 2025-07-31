package maestroFlowBuilder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testAutomationSuiteGO/internal/logger"
	"testAutomationSuiteGO/internal/testingToolkit"
)

type MaestroFlow struct {
	AppID string
	Steps []map[string]interface{}
}

type FlowBuilder struct {
	appID string
	steps []map[string]interface{}
}

func NewFlowBuilder(appID string) *FlowBuilder {
	return &FlowBuilder{appID: appID}
}

func (f *FlowBuilder) LaunchApp() *FlowBuilder {
	f.steps = append(f.steps, map[string]interface{}{
		"launchApp": nil,
	})
	return f
}

func (f *FlowBuilder) TapOn(element string) *FlowBuilder {
	f.steps = append(f.steps, map[string]interface{}{
		"tapOn": element,
	})
	return f
}

func (f *FlowBuilder) TapOnConnectedToggle() *FlowBuilder {
	f.steps = append(f.steps, map[string]interface{}{
		"tapOn": "Connected",
	})
	return f
}

func (f *FlowBuilder) TapOnPoint(xPercent, yPercent int) *FlowBuilder {
	f.steps = append(f.steps, map[string]interface{}{
		"tapOn": map[string]interface{}{
			"point": fmt.Sprintf("%d%%,%d%%", xPercent, yPercent),
		},
	})
	return f
}

func (f *FlowBuilder) InputText(text string) *FlowBuilder {
	f.steps = append(f.steps, map[string]interface{}{
		"inputText": text,
	})
	return f
}

func (f *FlowBuilder) AssertVisible(element string) *FlowBuilder {
	f.steps = append(f.steps, map[string]interface{}{
		"assertVisible": element,
	})
	return f
}

func (f *FlowBuilder) Swipe(direction string) *FlowBuilder {
	f.steps = append(f.steps, map[string]interface{}{
		"swipe": map[string]string{"direction": direction},
	})
	return f
}

func (f *FlowBuilder) WaitForAnimationToEnd() *FlowBuilder {
	f.steps = append(f.steps, map[string]interface{}{
		"waitForAnimationToEnd": nil,
	})
	return f
}

func (f *FlowBuilder) StopApp(appID string) *FlowBuilder {
	f.steps = append(f.steps, map[string]interface{}{
		"stopApp": appID,
	})
	return f
}

func (f *FlowBuilder) Build() *MaestroFlow {
	return &MaestroFlow{
		AppID: f.appID,
		Steps: f.steps,
	}
}

func (f *FlowBuilder) Save(filename string) error {
	flow := f.Build()

	var content strings.Builder
	content.WriteString(fmt.Sprintf("appId: %s\n---\n", flow.AppID))

	for _, step := range flow.Steps {
		for key, value := range step {
			switch v := value.(type) {
			case string:
				content.WriteString(fmt.Sprintf("- %s: \"%s\"\n", key, v))
			case map[string]interface{}:
				content.WriteString(fmt.Sprintf("- %s:\n", key))
				for k, val := range v {
					content.WriteString(fmt.Sprintf("    %s: %s\n", k, val))
				}
			case nil:
				content.WriteString(fmt.Sprintf("- %s\n", key))
			default:
				content.WriteString(fmt.Sprintf("- %s: %v\n", key, v))
			}
		}
	}

	return os.WriteFile(filename, []byte(content.String()), 0644)
}

func RunMaestroFlow(filePath, debugOutputPath string) (string, error) {
	if err := os.MkdirAll(debugOutputPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create debug output directory: %v", err)
	}
	cmd := exec.Command("maestro", "test", "--debug-output", debugOutputPath, filePath)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func MaestroOutputCleanUp(debugOutputPath, tcNumber string) {
	whereMaestroDataAppearsDir := debugOutputPath + "/.maestro/tests"
	dir, err := os.Open(whereMaestroDataAppearsDir)
	if err != nil {
		logger.Log("Error opening directory: "+err.Error(), tcNumber)
		return
	}
	defer dir.Close()

	folders, err := dir.Readdir(-1)
	if err != nil {
		logger.Log("Error reading directory: "+err.Error(), tcNumber)
		return
	}

	if len(folders) == 0 {
		logger.Log("No folders found in directory", tcNumber)
		return
	}
	folderName := folders[0].Name()
	sourceDir := filepath.Join(whereMaestroDataAppearsDir, folderName)

	sourceFolder, err := os.Open(sourceDir)
	if err != nil {
		logger.Log("Error opening source directory: "+err.Error(), tcNumber)
		return
	}
	defer sourceFolder.Close()

	files, err := sourceFolder.Readdir(-1)
	if err != nil {
		logger.Log("Error reading source directory: "+err.Error(), tcNumber)
		return
	}

	for _, file := range files {
		sourcePath := filepath.Join(sourceDir, file.Name())
		destPath := filepath.Join(debugOutputPath, file.Name())

		err := os.Rename(sourcePath, destPath)
		if err != nil {
			logger.Log("Error moving file "+file.Name()+": "+err.Error(), tcNumber)
			continue
		}
	}

	testingToolkit.DeleteFolder(debugOutputPath + "/.maestro")
}
