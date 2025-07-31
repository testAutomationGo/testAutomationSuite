package reporting

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"testAutomationSuiteGO/internal/logger"
	"testAutomationSuiteGO/internal/testRunParameters"
	"testAutomationSuiteGO/internal/testingToolkit"
	"time"
)

type TestCase struct {
	TcNumber          string            `json:"tcNumber"`
	Title             string            `json:"title"`
	Description       string            `json:"description"`
	Steps             map[string]string `json:"steps"`
	ExpectedOutcome   string            `json:"ExpectedOutcome"`
	TestData          template.HTML     `json:"TestData"`
	AccountData       string            `json:"accountData"`
	ActualOutput      template.HTML     `json:"ActualOutput"`
	FailureStatement  string            `json:"FailureStatement"`
	Result            string            `json:"Result"`
	DateExecuted      string            `json:"DateExecuted"`
	StartTime         string            `json:"StartTime"`
	EndTime           string            `json:"EndTime"`
	MillisecondsTaken string            `json:"MillisecondsTaken"`
	TestSector        string            `json:"TestSector"`
	Notes             string            `json:"Notes"`
	LogEntries        []string          `json:"LogEntries"`
}

type TestResults struct {
	TestCases          []TestCase            `json:"testCases"`
	NumberOfTests      string                `json:"NumberOfTests"`
	NumberOfTestsRan   string                `json:"NumberOfTestsRan"`
	TestsPassed        string                `json:"TestsPassed"`
	TestsFailed        string                `json:"TestsFailed"`
	ExecutionStartTime string                `json:"ExecutionStartTime"`
	ExecutionEndTime   string                `json:"ExecutionEndTime"`
	ExecutionDate      string                `json:"ExecutionDate"`
	ENV                string                `json:"ENV"`
	GroupedTests       map[string][]TestCase `json:"groupedTests"`
}

func UploadAndGenerateSignedURL() string {
	return "complete"
}

func GenerateHTMLReporter(resultsFile string) {
	err := waitForFile(resultsFile, 30*time.Second)
	if err != nil {
		log.Fatalf("Error waiting for results file: %v", err)
	}

	jsonFile, err := os.ReadFile(resultsFile)
	if err != nil {
		log.Fatal(err)
	}

	var results TestResults
	err = json.Unmarshal(jsonFile, &results)
	if err != nil {
		log.Printf("Warning: Error parsing JSON: %v", err)
	}

	logger.RetrieveLogDataFromLogFile(testRunParameters.GetResultsFolderPath() + "/logs.txt")
	logData := logger.GetLogData()

	associateLogData(&results, logData)

	results = ProcessTestResults(results.TestCases)

	tmpl := template.Must(template.New("results").Funcs(template.FuncMap{
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}).Parse(htmlTemplate))

	outputFile := filepath.Join(filepath.Dir(resultsFile), "test_results.html")
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = tmpl.Execute(file, results)
	if err != nil {
		log.Fatal(err)
	}
}

func associateLogData(results *TestResults, logData logger.LogData) {
	for i, testCase := range results.TestCases {
		if logs, exists := logData.Entries[testCase.TcNumber]; exists {
			results.TestCases[i].LogEntries = logs
		} else {
			results.TestCases[i].LogEntries = []string{}
		}
	}
}

func waitForFile(filename string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		fileInfo, err := os.Stat(filename)
		if err == nil {
			size := fileInfo.Size()
			time.Sleep(100 * time.Millisecond)
			newFileInfo, err := os.Stat(filename)
			if err != nil {
				return err
			}
			if newFileInfo.Size() == size {
				return nil
			}
			continue
		}
		if !os.IsNotExist(err) {
			return err
		}
		if time.Now().After(deadline) {
			return os.ErrNotExist
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func Format(title, value string) string {
	return "<li>" + title + ": " + value + "</li>"
}

func Break() string {
	return "<br>"
}

func ProcessTestResults(testCases []TestCase) TestResults {
	totalTests := testingToolkit.ConvertIntToString(len(testCases))
	testsRan := 0
	testsPassed := 0
	testsFailed := 0

	for _, testCase := range testCases {
		if testCase.Result != "" {
			testsRan++
			switch testCase.Result {
			case "Passed":
				testsPassed++
			case "Failed":
				testsFailed++
			}
		}
	}

	return TestResults{
		NumberOfTests:    totalTests,
		NumberOfTestsRan: testingToolkit.ConvertIntToString(testsRan),
		TestsPassed:      testingToolkit.ConvertIntToString(testsPassed),
		TestsFailed:      testingToolkit.ConvertIntToString(testsFailed),
		ExecutionDate:    time.Now().Format("2006-01-02 15:04:05"),
		ENV:              testRunParameters.GetEnvName(),
		TestCases:        testCases,
		GroupedTests:     GroupTestCasesBySector(testCases),
	}
}

func GroupTestCasesBySector(testCases []TestCase) map[string][]TestCase {
	grouped := make(map[string][]TestCase)

	for _, testCase := range testCases {
		sector := testCase.TestSector
		if sector == "" {
			sector = "Uncategorized"
		}
		grouped[sector] = append(grouped[sector], testCase)
	}

	return grouped
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Test Results</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        .test-category {
            margin-bottom: 20px;
        }
        .test-category h2 {
            background-color: #f4f4f4;
            padding: 10px;
            cursor: pointer;
            text-decoration: underline;
        }
        .test-case {
            border: 1px solid #ddd;
            margin-bottom: 10px;
            padding: 10px;
        }
        .test-case h3 {
            margin-top: 0;
            cursor: pointer;
        }
        .test-details {
            display: none;
        }
        .passed {
            color: green;
        }
        .failed {
            color: red;
        }
        .summary {
            margin-bottom: 20px;
            font-weight: bold;
        }
        .not-run {
            color: orange;
        }
    </style>
</head>
<body>
    <h1>Test Results</h1>
    <div class="summary">
        <p>Total Tests: {{if .NumberOfTests}}{{.NumberOfTests}}{{else}}N/A{{end}}</p>
        <p>Tests Ran: {{if .NumberOfTestsRan}}{{.NumberOfTestsRan}}{{else}}N/A{{end}}</p>
        <p style="color: green;">Tests Passed: {{if .TestsPassed}}{{.TestsPassed}}{{else}}N/A{{end}}</p>
        <p style="color: red;">Tests Failed: {{if .TestsFailed}}{{.TestsFailed}}{{else}}N/A{{end}}</p>
        <p>Execution Date: {{if .ExecutionDate}}{{.ExecutionDate}}{{else}}N/A{{end}}</p>
        <p>Environment: {{if .ENV}}{{.ENV}}{{else}}N/A{{end}}</p>
        <p>Click On the Below Links To View The Results</p>
    </div>

    {{range $sector, $tests := .GroupedTests}}
        {{template "testCategory" dict "Name" $sector "TestCases" $tests}}
    {{end}}

    <script>
        function toggleDetails(id) {
            var details = document.getElementById(id);
            if (details.style.display === "none") {
                details.style.display = "block";
            } else {
                details.style.display = "none";
            }
        }
    </script>
</body>
</html>

{{define "testCategory"}}
    {{- $hasFailed := false -}}
    {{- $hasPassed := false -}}
    
    {{- range .TestCases -}}
        {{- if eq .Result "Failed" -}}
            {{- $hasFailed = true -}}
        {{- else if eq .Result "Passed" -}}
            {{- $hasPassed = true -}}
        {{- end -}}
    {{- end -}}

    <div class="test-category">
        <h2 onclick="toggleDetails('{{.Name}}')" 
            class="{{if eq (len .TestCases) 0}}not-run{{end}}"
            {{if eq (len .TestCases) 0}}style="color: orange;"
            {{else if $hasFailed}}style="color: red;"
            {{else if $hasPassed}}style="color: green;"{{end}}>
            {{.Name}} {{if eq (len .TestCases) 0}}(Not Run){{end}}
        </h2>
        <div id="{{.Name}}" style="display: none;">
            {{if gt (len .TestCases) 0}}
                {{range .TestCases}}
                    {{template "testCase" .}}
                {{end}}
            {{else}}
                <p>No tests were run in this category.</p>
            {{end}}
        </div>
    </div>
{{end}}

{{define "testCase"}}
<div class="test-case">
    <h3 onclick="toggleDetails('{{.TcNumber}}')" class="{{if eq .Result "Passed"}}passed{{else}}failed{{end}}">
        {{.TcNumber}}: {{.Title}} ({{.Result}})
    </h3>
    <div id="{{.TcNumber}}" class="test-details">
        <p><strong>Description:</strong> {{.Description}}</p>
        <p><strong>Steps:</strong></p>
        <ul>
            {{range $step, $desc := .Steps}}
                <li>{{$step}}: {{$desc}}</li>
            {{end}}
        </ul>
        <p><strong>Expected Outcome:</strong> {{.ExpectedOutcome}}</p>
        <p><strong>Test Data:</strong> <ul>{{.TestData}}<li>Account: {{.AccountData}}</li></ul></p>
        <p><strong>Actual Output:</strong> <ul>{{.ActualOutput}}</ul></p>
		{{if eq .Result "Failed"}}
		    <p><strong>Failure Statement:</strong> {{.FailureStatement}}</p>
		{{end}}
        <p><strong>Date Executed:</strong> {{.DateExecuted}}</p>
        <p><strong>Start Time:</strong> {{.StartTime}}</p>
        <p><strong>End Time:</strong> {{.EndTime}}</p>
        <p><strong>This Test Case Time To Complete:</strong> {{.MillisecondsTaken}}</p>
        {{if .Notes}}
            <p><strong>Notes:</strong> {{.Notes}}</p>
        {{end}}
        <p><strong>Log Entries:</strong></p>
        {{if .LogEntries}}
            <ul>
                {{range .LogEntries}}
                    <li>{{.}}</li>
                {{end}}
            </ul>
        {{else}}    
            <p>No log entries found for this test case.</p>
        {{end}}
    </div>
</div>
{{end}}
`
