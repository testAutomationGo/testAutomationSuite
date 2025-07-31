package testingToolkit

import (
	"bufio"
	"encoding/json"
	"log"
	"path/filepath"

	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func GetENVVariable(fullFilePath, variableName string) string {
	file, err := os.Open(fullFilePath)
	if err != nil {
		log.Println("Error opening ENV variable file: ", err)
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, variableName+"=") {
			return strings.TrimPrefix(line, variableName+"=")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error reading env variable file: ", err)
		return ""
	}
	return ""
}

func PrintStringToFile(stringToFile, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err.Error())
		return
	}
	defer file.Close()
	_, err = file.WriteString(stringToFile)
	if err != nil {
		fmt.Println("Error writing to file:", err.Error())
		return
	}
}

func AppendStringToFile(stringToFile, filePath string) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err.Error())
		return
	}
	defer file.Close()
	_, err = file.WriteString(stringToFile)
	if err != nil {
		fmt.Println("Error appending to file:", err.Error())
		return
	}
}

func DeleteFile(filePath string) {
	err := os.Remove(filePath)
	if err != nil {
		fmt.Println("Error deleting file:", err.Error())
		return
	}
}

func DeleteFolder(folderPath string) {
	err := os.RemoveAll(folderPath)
	if err != nil {
		fmt.Println("Error deleting folder:", err.Error())
		return
	}
}

func CurrPath() string {
	currDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err.Error())
		return ""
	}
	return currDir
}

func LocalVariablePath() string {
	projectPath := CurrPath()
	if len(os.Args) < 2 {
		return projectPath + "/testData/localTestingVariables"
	}
	return projectPath + "/testData/localTestingVariables"
}

func SubstringToSpecific(base, cutoff string) string {
	index := strings.Index(base, cutoff)
	if index == -1 {
		return base
	}
	return base[:index+len(cutoff)]
}

func GetFunctionName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	pos := strings.LastIndex(fn.Name(), ".")
	if pos == -1 {
		return fn.Name()
	}
	funcName := fn.Name()[pos+1:]
	return funcName
}

func ConvertStringToInt(s string) int {
	num, err := strconv.Atoi(s)
	if err != nil {
		fmt.Println("Error converting string to int: " + err.Error())
		return 0
	}
	return num
}

func ConvertIntToString(i int) string {
	return strconv.Itoa(i)
}

func ConvertBoolToString(b bool) string {
	return strconv.FormatBool(b)
}

func ConvertInt64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func ConvertFloat64ToString(f float64, precision int) string {
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, f)
}

func ConvertStringToInt64(s string) int64 {
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

func CurrentDate() string {
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02")
	return formattedTime
}

func CurrentTimeForLogging() string {
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")
	return formattedTime
}

func CurrentTimeForLoggingWithMS() string {
	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05.000")
	formattedTime = strings.ReplaceAll(formattedTime, " ", "_")
	formattedTime = strings.ReplaceAll(formattedTime, ":", "")
	return formattedTime
}

func TimeDifferenceForLoggingWithMSInMS(startTime, endTime string) (int64, error) {
	layout := "2006-01-02_150405.000"

	start, err := time.Parse(layout, startTime)
	if err != nil {
		return 0, fmt.Errorf("error parsing start time: %v", err)
	}

	end, err := time.Parse(layout, endTime)
	if err != nil {
		return 0, fmt.Errorf("error parsing end time: %v", err)
	}

	duration := end.Sub(start)
	return duration.Milliseconds(), nil
}

func CurrentTimeForNaming() string {

	currentTime := time.Now()
	formattedTime := currentTime.Format("20060102_150405")
	formattedTime = strings.ReplaceAll(formattedTime, "_", "")

	return formattedTime
}

func CurrentTimeForNamingWithMS() string {
	currentTime := time.Now()
	stringAfyCurrentTime := currentTime.String()
	cutOfEnd := stringAfyCurrentTime[:23]
	removeDots := strings.ReplaceAll(cutOfEnd, ".", "")
	removeSpaces := strings.ReplaceAll(removeDots, " ", "")
	removeDashes := strings.ReplaceAll(removeSpaces, "-", "")
	properTime := strings.ReplaceAll(removeDashes, ":", "")
	return properTime
}

func ReadableCurrentTimeWithMS() string {
	currentTime := time.Now()
	stringAfyCurrentTime := currentTime.String()
	cutOfEnd := stringAfyCurrentTime[:23]
	return cutOfEnd
}

func CurrentTimeForTimeDurationWithMS() string {
	currentTime := time.Now()
	stringAfyCurrentTime := currentTime.String()
	cutOfEnd := stringAfyCurrentTime[:23]
	removeDots := strings.ReplaceAll(cutOfEnd, ":", "")
	removeSpaces := strings.ReplaceAll(removeDots, " ", "")
	removeDashes := strings.ReplaceAll(removeSpaces, "-", "")
	properTime := strings.ReplaceAll(removeDashes, ":", "")
	return properTime
}

func CreateFolderWithTodaysDate() {
	if _, err := os.Stat("reporting"); os.IsNotExist(err) {
		errDir := os.Mkdir("reporting", 0755)
		if errDir != nil {
			fmt.Println("Error creating reporting folder:", errDir.Error())
			return
		}
	}
	todaysDate := time.Now().Format("01022006")
	if _, err := os.Stat("reporting/" + todaysDate); os.IsNotExist(err) {
		errDir := os.Mkdir("reporting/"+todaysDate, 0755)
		if errDir != nil {
			fmt.Println("Error creating folder for today's date:", err.Error())
			return
		}
	}
}

func DelayMilliseconds(milliseconds int) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

func DelaySeconds(seconds int) {
	time.Sleep(time.Duration(seconds) * time.Second)
}

func DelayMinutes(minutes int) {
	time.Sleep(time.Duration(minutes) * time.Minute)
}

func RateLimitDelay() {
	time.Sleep(500 * time.Millisecond)
}

func VerifyStringIsAnInt(s string) bool {
	_, err := strconv.Atoi(s)
	if err == nil {
		return true
	} else {
		return false
	}
}

func VerifyFolderIsPresent(folderPath string) bool {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func VerifyFileIsPresent(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func ListFilesInFolder(folderPath string) []string {
	dir, err := os.Open(folderPath)
	if err != nil {
		return nil
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil
	}
	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}
	return fileNames
}

func DoesPathExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func CreateFolder(folderPath string, folderName string) {
	err := os.MkdirAll(folderPath+"/"+folderName, 0755)
	if err != nil {
		fmt.Println("Error creating folder:", err.Error())
		return
	}
}

func ReadTCNumbersFiles(resultsFolder, tcNumber string) []string {
	var resultsOfContent []string
	dir := resultsFolder
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("Error reading directory:", err.Error())
		return nil
	}
	var filesThatContainTCNumber []string
	for _, file := range files {
		if strings.Contains(file.Name(), tcNumber) {
			filesThatContainTCNumber = append(filesThatContainTCNumber, file.Name())
		}
	}
	for _, file := range filesThatContainTCNumber {
		fileText, err := os.ReadFile(dir + "/" + file)
		if err != nil {
			fmt.Println("Error reading file:", err.Error())
			return nil
		}
		fileContent := string(fileText)
		resultsOfContent = append(resultsOfContent, fileContent)
	}
	return resultsOfContent
}

func ReadFileContentToString(results string) string {
	fileText, err := os.ReadFile(results)
	if err != nil {
		fmt.Println("Error reading file:", err.Error())
		return ""
	}
	fileContent := string(fileText)
	return fileContent
}

func IsStringNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func RemoveDuplicateFolders(path string) string {

	if path == "" {
		return path
	}

	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))

	var cleanParts []string
	for _, part := range parts {
		if part != "" {
			cleanParts = append(cleanParts, part)
		}
	}

	if len(cleanParts) <= 1 {
		return path
	}

	var result []string
	result = append(result, cleanParts[0])

	for i := 1; i < len(cleanParts); i++ {
		if cleanParts[i] != cleanParts[i-1] {
			result = append(result, cleanParts[i])
		}
	}

	cleanedPath := strings.Join(result, string(filepath.Separator))

	if len(parts) > 0 && strings.Contains(parts[0], ":") {
		return cleanedPath
	}

	if strings.HasPrefix(path, string(filepath.Separator)) {
		cleanedPath = string(filepath.Separator) + cleanedPath
	}

	return cleanedPath
}

type JSONGrabber struct {
	data map[string]interface{}
}

func NewJSONGrabberFromString(jsonStr string) (*JSONGrabber, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, err
	}
	return &JSONGrabber{data: result}, nil
}

func NewJSONGrabber(filename string) (*JSONGrabber, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &JSONGrabber{data: result}, nil
}

func (jg *JSONGrabber) GetValue(key string) (interface{}, bool) {
	value, exists := jg.data[key]
	return value, exists
}

func (jg *JSONGrabber) GetAllKeys() []string {
	keys := make([]string, 0, len(jg.data))
	for key := range jg.data {
		keys = append(keys, key)
	}
	return keys
}

func (jg *JSONGrabber) SearchKeys(pattern string) map[string]interface{} {
	results := make(map[string]interface{})
	for key, value := range jg.data {
		if strings.Contains(strings.ToLower(key), strings.ToLower(pattern)) {
			results[key] = value
		}
	}
	return results
}

func (jg *JSONGrabber) GetNestedValue(path string) (interface{}, bool) {
	keys := strings.Split(path, ".")
	var current interface{} = jg.data

	for _, key := range keys {
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[key]; exists {
				current = val
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	return current, true
}

func GetMaxInt64(times []int64) int64 {
	if len(times) == 0 {
		return 0
	}

	max := times[0]
	for _, time := range times {
		if time > max {
			max = time
		}
	}
	return max
}

func GetMinInt64(times []int64) int64 {
	if len(times) == 0 {
		return 0
	}

	min := times[0]
	for _, time := range times {
		if time < min {
			min = time
		}
	}
	return min
}

func GetAverageInt64(times []int64) float64 {
	if len(times) == 0 {
		return 0
	}

	var sum int64
	for _, time := range times {
		sum += time
	}
	return float64(sum) / float64(len(times))
}

func MaskIP(ip string) string {
	if len(ip) < 2 {
		return ip
	}

	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return ip
	}

	firstOctet := parts[0]
	if len(firstOctet) < 2 {
		return ip
	}

	maskedFirstOctet := firstOctet[:2] + strings.Repeat("X", len(firstOctet)-2)

	return maskedFirstOctet + ".X.X.X"
}
