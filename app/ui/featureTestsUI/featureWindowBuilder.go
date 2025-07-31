package featureTestsUI

import (
	"log"
	"os"
	"reflect"
	"testAutomationSuiteGO/featureTesting/featureTestingFunctionRegistry"
	"testAutomationSuiteGO/internal/testingToolkit"

	"fyne.io/fyne/v2"
)

func GenerateFeatureTestTabs() []FeatureTestTab {
	dirPath := testingToolkit.CurrPath() + "/featureTesting"
	features, err := os.ReadDir(dirPath)
	if err != nil {
		log.Println("Error reading feature tests directory:", err)
		return nil
	}
	var tabs []FeatureTestTab

	for _, feature := range features {
		if feature.IsDir() {
			object, exists := GetFeatureCanvasObject(feature.Name() + ".GenerateWindowContent")
			if !exists {
				continue
			}
			tab := FeatureTestTab{
				name:    feature.Name(),
				content: object,
			}
			tabs = append(tabs, tab)
		}
	}
	return tabs
}

func GetFeatureCanvasObject(featureName string) (fyne.CanvasObject, bool) {
	funcValue, exists := featureTestingFunctionRegistry.FunctionRegistry[featureName]
	if !exists {
		return nil, false
	}

	if !funcValue.IsValid() {
		return nil, false
	}

	if funcValue.Kind() != reflect.Func {
		return nil, false
	}

	funcType := funcValue.Type()
	if funcType.NumIn() != 0 {
		return nil, false
	}

	results := funcValue.Call([]reflect.Value{})

	if len(results) == 0 {
		return nil, false
	}

	result := results[0].Interface()
	canvasObj, ok := result.(fyne.CanvasObject)
	if !ok {
		return nil, false
	}

	return canvasObj, true
}
