package utils;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileReader;
import java.util.ArrayList;
import java.util.List;

import org.json.JSONObject;

import org.json.JSONArray;

public class NotionJSONs {
    //This class is to store all the JSONs that are used in the NotionResultsReporting class, these JSONs are so long they will need their own file.
    
    ReusableTasks rt = new ReusableTasks();

    public String resutltsRunPagePrefix(String currentMonthReportingPageID) {
        //This JSON part hold the parent section of payload.
        return "{\r\n" +
                "  \"parent\": {\r\n" +
                "    \"type\": \"page_id\",\r\n" +
                "    \"page_id\": \"" + currentMonthReportingPageID + "\"\r\n" +
                "  },\r\n";
    }
    
    public String runResultsOverallDataJSON(String testExecutionTitle, String numberOfTotalTests, String passedTests, String failedTests, String dateExecuted, String timeExecuted, String totalTime,
        JSONObject fullTestCaseExecutedDetails, String resultsFolderFullPath) {
        String[] filesList = new File(resultsFolderFullPath).list();
        String resultsXML = "Commandlinetest.xml";
        String printStreamLog = "";
        String loggedMessagesLogFile = "loggedMessagesLogFile.txt";
        String parameterJson = "parameterJson.txt";
        String fileNamesList = "folderAndFileEndPathsLogList.txt";
        for (int i = 0;i<filesList.length;i++) {
            if (filesList[i].contains("PrintStreamLog")) {
                printStreamLog = filesList[i];
            }
        }
        String runResultsOverallDataJSON =
                "  \"properties\": {\r\n" +
                "    \"title\": {\r\n" +
                "      \"id\": \"title\",\r\n" +
                "      \"type\": \"title\",\r\n" +
                "      \"title\": [\r\n" +
                "        {\r\n" +
                "          \"type\": \"text\",\r\n" +
                "          \"text\": {\r\n" +
                "            \"content\": \"" + testExecutionTitle + "\",\r\n" +
                "            \"link\": null\r\n" +
                "          }\r\n" +
                "        }\r\n" +
                "      ]\r\n" +
                "    }\r\n" +
                "  },\r\n" + 
                "  \"children\": [\r\n" + 
                childObjectBlock("Total Number Of Tests:", numberOfTotalTests, "true", "default") +
                ",\r\n" +
                childObjectBlock("Pass:", passedTests, "true", "green") +
                ",\r\n" +
                childObjectBlock("Failed:", failedTests, "true", "red") +
                ",\r\n" +
                childObjectBlock("Date Executed:", dateExecuted, "false", "default") +
                ",\r\n" +
                childObjectBlock("Time Executed:", timeExecuted + " GMT", "false", "default") +
                ",\r\n" +
                childObjectBlock("Total Time:", totalTime, "false", "default") +
                ",\r\n" +
                spacerBlock() +
                ",\r\n" +
                generateTestCaseBlocks(fullTestCaseExecutedDetails, resultsFolderFullPath) +
                ",\r\n" +
                spacerBlock() +
                ",\r\n" +
                fileBlock(resultsXML, resultsFolderFullPath) +
                ",\r\n" +
                toggleBlockTooLong(printStreamLog, resultsFolderFullPath) +
                ",\r\n" +
                toggleBlockTooLong(loggedMessagesLogFile, resultsFolderFullPath) +
                ",\r\n" +
                fileBlock(parameterJson, resultsFolderFullPath) +
                ",\r\n" +
                fileBlock(fileNamesList, resultsFolderFullPath) +
                "\r\n" +
                "  ]\r\n" +
                "}";
        return runResultsOverallDataJSON;
    }

    String toggleBlockTooLong(String fileName, String resultsFolderFullPath) {
        //Toggles only allow for 100 lines of text to be hidden in the toggle.
        ArrayList<String> fileLines = new ArrayList<String>();
        try {
            File file = new File(resultsFolderFullPath + "/" + fileName);
            BufferedReader br = new BufferedReader(new FileReader(file));
            String st;
            while ((st = br.readLine()) != null) {
                fileLines.add(st);
            }
            br.close();
        }
        catch (Exception e) {
            System.out.println("--" + e.toString());
        }
        int howManyNewLists = (fileLines.size() + 99/100);
        ArrayList<List<String>> sublists = new ArrayList<>(howManyNewLists);

        // Create sublists
        for (int i = 0; i < howManyNewLists; i++) {
            int start = i * 100;
            int end = Math.min(start + 100, fileLines.size()); // Ensure we don't go past the end of the list
            if (start < fileLines.size()) {
                sublists.add(new ArrayList<>(fileLines.subList(start, end)));
            }
        }
        String fileToggleBlock = "";
        for (int i = 0; i <sublists.size();i++) {
            if (sublists.size() == 1) {
                fileToggleBlock = fileToggleBlock + fileJson(fileName, (ArrayList<String>) sublists.get(i));
                break;
            }
            else {
                fileToggleBlock = fileToggleBlock + fileJson("Part " + (i + 1) + " of " + fileName, (ArrayList<String>) sublists.get(i));
            }
            if (i != sublists.size() - 1) {
                fileToggleBlock = fileToggleBlock + ",\r\n";
            }
        }
        return fileToggleBlock;
    }

    String fileBlock(String fileName, String resultsFolderFullPath) {
        ArrayList<String> fileLines = new ArrayList<String>();
        try {
            File file = new File(resultsFolderFullPath + "/" + fileName);
            BufferedReader br = new BufferedReader(new FileReader(file));
            String st;
            while ((st = br.readLine()) != null) {
                fileLines.add(st);
            }
            br.close();
        }
        catch (Exception e) {
            System.out.println("--" + e.toString());
        }
        String fileToggleBlock = fileJson(fileName, fileLines);
        return fileToggleBlock;
    }

    public String fileLinesBlock(ArrayList<String> fileLines) {
        String fileLinesBlock = "";
        for (int i = 0;i<fileLines.size();i++) {
            String fileLine = fileLines.get(i);
            if (fileLine.length() > 1900) {
                fileLine = fileLine.substring(0, 1900);
                
            }
            fileLine = fileLine.replace("\"", "\\\"");
            fileLinesBlock = fileLinesBlock + 
                "              {\r\n" +
                "                \"type\": \"text\",\r\n" +
                "                \"text\": {\r\n" +
                "                  \"content\": \"   " + fileLine + "\\n\"\r\n" +
                "                }\r\n" +
                "              }\r\n";
            if (i != fileLines.size() - 1) {
                fileLinesBlock = fileLinesBlock + ",\r\n";
            }
        }
        return fileLinesBlock;
    }

    public String fileJson(String fileName, ArrayList<String> fileLines) {
        String fileJson = 
                "    {\r\n" +
                "      \"object\": \"block\",\r\n" +
                "      \"type\": \"toggle\",\r\n" +
                "      \"toggle\": {\r\n" +
                "        \"rich_text\": [\r\n" +
                "          {\r\n" +
                "            \"type\": \"text\",\r\n" +
                "            \"text\": {\r\n" +
                "              \"content\": \"" + fileName + "\",\r\n" +
                "              \"link\": null\r\n" +
                "            },\r\n" +
                "            \"annotations\": {\r\n" +
                "              \"underline\": true,\r\n" +
                "              \"bold\": true,\r\n" +
                "              \"color\": \"default\"\r\n" +
                "            }\r\n" +
                "          }\r\n" +
                "        ],\r\n" +
                "        \"children\": [\r\n" +
                "          {\r\n" +
                "            \"object\": \"block\",\r\n" +
                "            \"type\": \"paragraph\",\r\n" +
                "            \"paragraph\": {\r\n" +
                "              \"rich_text\": [\r\n" +
                fileLinesBlock(fileLines) +
                "              ]\r\n" +
                "            }\r\n" +
                "          }\r\n" +
                "        ]\r\n" +
                "      }\r\n" +
                "    }";
                return fileJson;
    }

    public String childObjectBlock(String contentString, String contentObjectData, String boldTrueOrFalse, String fontColor) {
        String objectBlock = 
                "    {\r\n" + 
                "      \"object\": \"block\",\r\n" + 
                "      \"type\": \"paragraph\",\r\n" + 
                "      \"paragraph\": {\r\n" + 
                "        \"rich_text\": [\r\n" + 
                "          {\r\n" + 
                "            \"type\": \"text\",\r\n" + 
                "            \"text\": {\r\n" + 
                "              \"content\": \"" + contentString + " " + contentObjectData + "\",\r\n" + 
                "              \"link\": null\r\n" + 
                "            },\r\n" +
                "            \"annotations\": {\r\n" +
                "              \"bold\": " + boldTrueOrFalse + ",\r\n" +
                "              \"color\": \"" + fontColor + "\"\r\n" +
                "            }\r\n" +
                "          }\r\n" + 
                "        ]\r\n" + 
                "      }\r\n" +
                "    }";
        return objectBlock;
    }

    public String spacerBlock() {
        return 
                "    {\r\n" +
                "      \"object\": \"block\",\r\n" +
                "      \"type\": \"paragraph\",\r\n" +
                "      \"paragraph\": {\r\n" +
                "        \"rich_text\": []\r\n" +
                "      }\r\n" +
                "    }";
    }

    public String generateTestCaseBlocks(JSONObject fullTestCaseDetailsObject, String resultsFolderFullPath) {
        String testCaseBlocks = "";
        JSONArray testCaseArray = fullTestCaseDetailsObject.getJSONArray("ResultsOfTests");
        String testCaseDataFromREGJSON = rt.readFile(rt.currPath() + "/RegressionJSONTestCases.json");
        JSONObject testCaseDataObjectFronREGJSON = new JSONObject(testCaseDataFromREGJSON);
        JSONArray testCaseDataArrayFromREGJSON = testCaseDataObjectFronREGJSON.getJSONArray("testCases"); 
        for (int i = 0;testCaseArray.length() > i;i++) {
            JSONObject testCaseResultObject = testCaseArray.getJSONObject(i);
            String tcNumber = testCaseResultObject.getString("tcNumber");
            try {tcNumber = tcNumber.replace("-", "");} catch(Exception e) {}
            String title = "";
            String passOrFail = "";
            String description = "";
            String expectedOutcome = "";
            String actualOutcome = "";
            JSONObject steps = new JSONObject();
            for (int j = 0;testCaseDataArrayFromREGJSON.length() > j;j++) {
                JSONObject testCaseDataObject = testCaseDataArrayFromREGJSON.getJSONObject(j);
                String regJsonTcNumber = testCaseDataObject.getString("tcNumber");
                if (regJsonTcNumber.equals(tcNumber)) {
                    title = testCaseDataObject.getString("title");
                    description = testCaseDataObject.getString("description");
                    passOrFail = testCaseResultObject.getString("status"); 
                    expectedOutcome = testCaseDataObject.getString("ExpectedOutcome");
                    actualOutcome = "Meets expected result. See log statements.";
                    if (passOrFail.equals("FAIL")) {
                        actualOutcome = testCaseResultObject.getString("butFound");
                    }
                    steps = testCaseDataObject.getJSONObject("Steps");
                    break;
                }
            }
            testCaseBlocks = testCaseBlocks + 
                    testCaseBlock(title, passOrFail, description, expectedOutcome, actualOutcome, steps, resultsFolderFullPath, testCaseResultObject, tcNumber);
            if (i != testCaseArray.length() - 1) {
                testCaseBlocks = testCaseBlocks + ",\r\n";
            }
        }
        return testCaseBlocks;
    }

    public String testCaseBlock(String title, String passOrFail, String description, String expectedOutcome, String actualOutcome, JSONObject steps, String resultsFolderFullPath, JSONObject testObject,String tcNumber) {
        String testCase = 
                "    {\r\n" + 
                "      \"object\": \"block\",\r\n" + 
                "      \"type\": \"toggle\",\r\n" + 
                "      \"toggle\": {\r\n" + 
                "        \"rich_text\": [\r\n" + 
                "          {\r\n" + 
                "            \"type\": \"text\",\r\n" + 
                "            \"text\": {\r\n" + 
                "              \"content\": \"" + tcNumber + ": " + title + "\",\r\n" +
                "              \"link\": null\r\n" + 
                "            },\r\n" + 
                "            \"annotations\": {\r\n" + 
                "              \"underline\": true,\r\n" + 
                "              \"bold\": true,\r\n" +
                "              \"color\": \"" + passOrFailFont(passOrFail) + "\"\r\n" +
                "            }\r\n" + 
                "          },\r\n" + 
                "          {\r\n" + 
                "            \"type\": \"text\",\r\n" + 
                "            \"text\": {\r\n" + 
                "              \"content\": \" " + passOrFail + "\"\r\n" + 
                "            },\r\n" +
                "            \"annotations\": {\r\n" + 
                "              \"color\": \"" + passOrFailFont(passOrFail) + "\",\r\n" +
                "              \"bold\": true\r\n" +
                "            }\r\n" +
                "          }\r\n" + 
                "        ],\r\n" +
                "        \"children\": [\r\n" +
                fullTestCaseText(description, expectedOutcome, actualOutcome, passOrFail, resultsFolderFullPath, steps, testObject, tcNumber) +
                "          ]\r\n" +
                "        }\r\n" +
                "     }";
        return testCase;
    }

    public String passOrFailFont(String passOrFail) {
        String font = "";
        if (passOrFail.equals("PASS")) {
            font = "green";
        }
        else if (passOrFail.equals("FAIL")) {
            font = "red";
        }
        return font;
    }

    public String fullTestCaseText(String description, String expectedOutCome, String actualOutcome, String passOrFail, String resultsFolderFullPath, JSONObject steps, JSONObject testObject, String tcNumber) {
        String testCaseText = 
                "        {\r\n" +
                "          \"object\": \"block\",\r\n" +
                "          \"type\": \"paragraph\",\r\n" +
                "          \"paragraph\": {\r\n" +
                "            \"rich_text\": [\r\n" +
                "              {\r\n" +
                "                \"type\": \"text\",\r\n" +
                "                \"text\": {\r\n" +
                "                  \"content\": \"Description: " + description + "\\n\\n\"\r\n" +
                "                }\r\n" +
                "              },\r\n" +
                "              {\r\n" +
                "                \"type\": \"text\",\r\n" +
                "                \"text\": {\r\n" +
                "                  \"content\": \"Expected Outcome: " + expectedOutCome + "\\n\\n\"\r\n" +
                "                }\r\n" +
                "              },\r\n" +
                "              {\r\n" +
                "                \"type\": \"text\",\r\n" +
                "                \"text\": {\r\n" +
                "                  \"content\": \"Actual Outcome: " + actualOutcomeText(actualOutcome, passOrFail, testObject) + "\\n\\n\"\r\n" +
                "                }\r\n" +
                "              },\r\n" +
                "              {\r\n" +
                "                \"type\": \"text\",\r\n" +
                "                \"text\": {\r\n" +
                "                  \"content\": \"Logged Messages: \\n\"\r\n" +
                "                }\r\n" +
                "              },\r\n" +
                loggedMessageText(resultsFolderFullPath, tcNumber) +
                "              {\r\n" +
                "                \"type\": \"text\",\r\n" +
                "                \"text\": {\r\n" +
                "                  \"content\": \"\\nSteps: \\n\"\r\n" +
                "                }\r\n" +
                "              },\r\n" +
                stepsText(steps) +
                "            ]\r\n" + 
                "          }\r\n" +
                "        }\r\n";
        return testCaseText;
    }

    public String actualOutcomeText(String actualOutcome, String passOrFail, JSONObject testObject) {
        String actualOutcomeText = "Meets expected result. See log statements.";
        if (passOrFail.equals("FAIL")) {
            try {
                actualOutcomeText = testObject.getString("butFound");
            } catch (Exception e) {
                System.out.println("--" + e.toString());
                actualOutcomeText = "Unable to get actual outcome from test case.";
                actualOutcome = "Test failed. Unable to get actual outcome from test case.";
            }
        }
        return actualOutcomeText;
    }

    public String loggedMessageText(String resultsFolderFullPath, String tcNumber) {
        String loggedMessagesBlocks = "";
        try {   
            ArrayList<String> fileLinesArray = new ArrayList<String>();
			BufferedReader br = new BufferedReader(
					new FileReader(new File(resultsFolderFullPath + "/loggedMessagesLogFile.txt")));
			String st;
			while ((st = br.readLine()) != null) {
				fileLinesArray.add(st);
			}
			int count = 1;
			for (int i = 0; i < fileLinesArray.size(); i++) {
				if (fileLinesArray.get(i).contains(tcNumber + ":")) {
					String message = fileLinesArray.get(i).substring(tcNumber.length() + 2);
                    if (message.length() > 1900) {
                        message = message.substring(0, 1900);
                    }
                    message = message.replace("\"", "\\\"");
                    loggedMessagesBlocks = loggedMessagesBlocks +  
                        "              {\r\n" +
                        "                \"type\": \"text\",\r\n" +
                        "                \"text\": {\r\n" +
                        "                  \"content\": \"   " + count + ". "+ message + "\\n\"\r\n" +
                        "                }\r\n" +
                        "              },\r\n";
					count++;
				}
			}
            br.close();
        }
        catch (Exception e) {
            System.out.println("--" + e.toString());
            loggedMessagesBlocks = "{},";
        }
        return loggedMessagesBlocks;
    }

    public String stepsText(JSONObject steps){
        String stepsText = "";
        for (int i = 1; i < steps.length() + 1; i++) {
            String step = steps.getString(String.valueOf(i));
            stepsText = stepsText + 
                "              {\r\n" +
                "                \"type\": \"text\",\r\n" +
                "                \"text\": {\r\n" +
                "                  \"content\": \"   " + i + ". "+ step + "\\n\"\r\n" +
                "                }\r\n" +
                "              }\r\n";
            if (i != steps.length()) {
                stepsText = stepsText + ",\r\n";
            }
        }
        return stepsText;
    }
}
