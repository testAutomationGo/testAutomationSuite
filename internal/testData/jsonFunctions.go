package testData

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	testingToolkit "testAutomationSuiteGO/internal/testingToolkit"
)

func CreateUniqueJSON() string {
	json := "{\"typeOfData\":\"json\",\"createdAt\":\"" + testingToolkit.CurrentTimeForNamingWithMS() + "\",\"message\":\"Hello, this is a json.\",\"data\":\"" + testingToolkit.GetAlNumString(10) + testingToolkit.GetNumericString(5) + "\"}"
	return json
}

func JsonStringToJson(jsonString string) interface{} {
	var result interface{}
	json.Unmarshal([]byte(jsonString), &result)
	return result
}

func ConvertJsonToString(jsonData interface{}) string {
	jsonString, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Println("Error converting JSON to string:", err.Error())
		return ""
	}
	return string(jsonString)
}

func SaveJsonToFile(jsonToFile map[string]interface{}, fileName, folder string) {
	var file *os.File
	var err error
	if folder == "" {
		file, err = os.Create(fileName + ".json")
	} else {
		file, err = os.Create(folder + "/" + fileName + ".json")
	}
	if err != nil {
		fmt.Println("Error creating file:", err.Error())
		return
	}
	defer file.Close()
	byteData, err := json.Marshal(jsonToFile)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err.Error())
		return
	}
	_, err = file.Write(byteData)
	if err != nil {
		fmt.Println("Error writing to file:", err.Error())
		return
	}
}

func MarshalJsonStringToObject(jsonString string) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err.Error())
		return nil
	}
	return result
}

func ConvertJsonFileToObject(fullFilePath string) []byte {
	file, err := os.Open(fullFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err.Error())
		return nil
	}
	defer file.Close()
	byteValue, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error reading file:", err.Error())
		return nil
	}
	var resultInterface map[string]interface{}
	err = json.Unmarshal(byteValue, &resultInterface)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err.Error())
		return nil
	}
	resultJson, err := json.Marshal(resultInterface)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err.Error())
		return nil
	}
	return resultJson
}
