package reporting

import (
	"html/template"
	"log"
	"os"
)

func GenerateAndUploadGenricHMTLSignedURL() string {

	htmlFileName := "test_results.html"

	return htmlFileName
}

type SimpleReportData struct {
	Title         string
	Environment   string
	ExecutionTime string
	Additional    string
}

func GenerateSimpleHTMLReport(reportFolderLocation, htmlFileName string, data SimpleReportData) string {
	tmpl := template.Must(template.New("simpleReport").Parse(simpleHTMLTemplate))

	file, err := os.Create(reportFolderLocation + "/" + htmlFileName)
	if err != nil {
		log.Fatal(err)
		return "Unable to create HTML file"
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		log.Fatal(err)
		return "Unable to execute HTML template"
	}

	log.Printf("HTML report generated successfully: %s", htmlFileName)
	return htmlFileName
}

const simpleHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
        }
        .report-header {
            background-color: #f4f4f4;
            padding: 15px;
            text-align: center;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .report-section {
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 8px;
            margin-bottom: 15px;
        }
    </style>
</head>
<body>
    <div class="report-header">
        <h1>{{.Title}}</h1>
    </div>
    <div class="report-section">
        <p><strong>Environment:</strong> {{.Environment}}</p>
        <p><strong>Execution Date/Time:</strong> {{.ExecutionTime}}</p>
    </div>
    <div class="report-section">
        <h2>Additional Data</h2>
        <p>{{.Additional}}</p>
    </div>
</body>
</html>
`
