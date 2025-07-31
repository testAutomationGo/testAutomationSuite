package ollamaInternal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type TestAnalyzer struct {
	modelName string
}

func NewTestAnalyzer() *TestAnalyzer {
	return &TestAnalyzer{
		modelName: "llama3.1:8b",
	}
}

func (ta *TestAnalyzer) AnalyzeTestReport(htmlContent string) (string, error) {
	fmt.Println("Analyzing test report with Ollama...")

	// Create a temporary file for the prompt to avoid command line length limits
	tempFile, err := os.CreateTemp("", "ollama_prompt_*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	prompt := fmt.Sprintf(`You are an expert test analyst. Analyze this HTML test report and create a clear executive summary for non-technical stakeholders based ONLY on the actual test results shown.

HTML Test Report:
%s

Please provide a professional executive summary in HTML format that includes:

1. **Overall Test Status** - Pass/Fail summary with actual metrics from the report
2. **Critical Issues** - ONLY list actual failures or problems found in the test results
3. **Performance Summary** - Actual timing and performance metrics from the results
4. **Business Impact** - What the actual results mean for the business
5. **Recommendations** - ONLY provide recommendations for tests that actually failed or had issues. Do NOT recommend re-running passed tests or suggest process improvements unless there are actual problems shown in the results.

IMPORTANT ANALYSIS RULES:
- Base your analysis ONLY on what the test results actually show
- If tests passed, report them as successful - do not suggest they need improvement
- Only provide recommendations for tests that actually failed or encountered errors  
- Do not make assumptions about test adequacy or coverage unless explicitly shown in the results
- Focus on facts from the actual test execution, not theoretical improvements

IMPORTANT: Format your response as clean HTML content using:
- <h2> tags for main section headers
- <h3> tags for subsections
- <p> tags for paragraphs
- <ul> and <li> tags for lists
- <strong> tags for emphasis
- <span class="status-pass"> for successful items
- <span class="status-fail"> for failed items
- <span class="status-warning"> for warnings

Do NOT include <html>, <head>, or <body> tags - just the content that will go inside the body.
Keep it professional and based strictly on the actual test results.`, htmlContent)

	_, err = tempFile.WriteString(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to write to temp file: %v", err)
	}
	tempFile.Close()

	cmd := exec.Command("ollama", "run", ta.modelName)

	promptContent, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read temp file: %v", err)
	}

	cmd.Stdin = strings.NewReader(string(promptContent))

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run Ollama: %v", err)
	}

	return string(output), nil
}

func (ta *TestAnalyzer) ProcessHTMLFile(filePath string) error {
	fmt.Printf("Reading HTML file: %s\n", filePath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read HTML file: %v", err)
	}

	aiContent, err := ta.AnalyzeTestReport(string(content))
	if err != nil {
		return fmt.Errorf("failed to analyze test report: %v", err)
	}

	htmlSummary := ta.createHTMLReport(strings.TrimSpace(aiContent), filePath)

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("EXECUTIVE SUMMARY GENERATED")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("HTML report created with professional styling!")

	outputFile := strings.TrimSuffix(filePath, ".html") + "_executive_summary.html"
	err = os.WriteFile(outputFile, []byte(htmlSummary), 0644)
	if err != nil {
		return fmt.Errorf("failed to save HTML summary: %v", err)
	}

	fmt.Printf("\nExecutive summary saved to: %s\n", outputFile)
	fmt.Printf("Open the file in your browser to view the formatted report.\n")

	return nil
}

func (ta *TestAnalyzer) createHTMLReport(aiContent, originalFile string) string {
	timestamp := time.Now().Format("January 2, 2006 at 3:04 PM")

	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Executive Test Summary</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
        }
        .header {
            border-bottom: 3px solid #007acc;
            padding-bottom: 20px;
            margin-bottom: 30px;
        }
        .header h1 {
            color: #007acc;
            margin: 0;
            font-size: 2.5em;
        }
        .meta-info {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 30px;
            border-left: 4px solid #007acc;
        }
        .meta-info p {
            margin: 5px 0;
            color: #666;
        }
        h2 {
            color: #007acc;
            border-bottom: 2px solid #e9ecef;
            padding-bottom: 10px;
            margin-top: 40px;
            margin-bottom: 20px;
        }
        h3 {
            color: #495057;
            margin-top: 25px;
            margin-bottom: 15px;
        }
        .status-pass {
            color: #28a745;
            font-weight: bold;
            background: #d4edda;
            padding: 2px 8px;
            border-radius: 4px;
        }
        .status-fail {
            color: #dc3545;
            font-weight: bold;
            background: #f8d7da;
            padding: 2px 8px;
            border-radius: 4px;
        }
        .status-warning {
            color: #ffc107;
            font-weight: bold;
            background: #fff3cd;
            padding: 2px 8px;
            border-radius: 4px;
        }
        ul {
            padding-left: 20px;
        }
        li {
            margin-bottom: 8px;
        }
        .highlight-box {
            background: #e3f2fd;
            border: 1px solid #bbdefb;
            border-radius: 5px;
            padding: 20px;
            margin: 20px 0;
        }
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #e9ecef;
            text-align: center;
            color: #666;
            font-size: 0.9em;
        }
        @media print {
            body {
                background-color: white;
            }
            .container {
                box-shadow: none;
                padding: 20px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Executive Test Summary</h1>
        </div>
        
        <div class="meta-info">
            <p><strong>Report Generated:</strong> %s</p>
            <p><strong>Source File:</strong> %s</p>
            <p><strong>Analysis Tool:</strong> Ollama AI Test Analyzer</p>
        </div>

        %s

        <div class="footer">
            <p>This executive summary was automatically generated from test results using AI analysis.</p>
        </div>
    </div>
</body>
</html>`

	return fmt.Sprintf(htmlTemplate, timestamp, filepath.Base(originalFile), aiContent)
}

func (ta *TestAnalyzer) CheckOllamaInstalled() error {
	cmd := exec.Command("ollama", "--version")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ollama not found. please install ollama first")
	}
	return nil
}

func (ta *TestAnalyzer) CheckModelExists() error {
	fmt.Printf("Checking if model '%s' is available...\n", ta.modelName)

	cmd := exec.Command("ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list Ollama models: %v", err)
	}

	if !strings.Contains(string(output), ta.modelName) {
		fmt.Printf("Model '%s' not found. Pulling model...\n", ta.modelName)
		pullCmd := exec.Command("ollama", "pull", ta.modelName)
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr

		err = pullCmd.Run()
		if err != nil {
			return fmt.Errorf("failed to pull model '%s': %v", ta.modelName, err)
		}
		fmt.Println("Model downloaded successfully!")
	} else {
		fmt.Printf("Model '%s' is ready!\n", ta.modelName)
	}

	return nil
}
