package apiFunctions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func PostFileRequest(url string, jwt string, fileName string, folderPath string) (int, string) {
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(folderPath + "/" + fileName)
	if errFile1 != nil {
		fmt.Println(errFile1)
		return 0, errFile1.Error()
	}
	defer file.Close()
	part1, _ := writer.CreateFormFile("file", filepath.Base(folderPath+"/"+fileName))
	_, err := io.Copy(part1, file)
	if err != nil {
		fmt.Println(err)
		return 0, err.Error()
	}

	err = writer.Close()
	if err != nil {
		fmt.Println(err)
		return 0, err.Error()
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return 0, err.Error()
	}
	req.Header.Add("Authorization", "Bearer "+jwt)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return 0, err.Error()
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return 0, err.Error()
	}

	return res.StatusCode, string(body)
}

func CurlPostFileRequest(url, jwt string, headers []string) (int, string) {

	args := []string{
		"--request", "POST",
		"--url", url,
		"--header", "Authorization: Bearer " + jwt,
	}

	for i := 0; i < len(headers); i++ {
		args = append(args, "--header", headers[i])
	}
	//args = append(args, "--form", "network="+network)
	cmd := exec.Command("curl", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error executing curl:", err)
		fmt.Println("Stderr:", stderr.String())
		return 0, err.Error()
	}

	responseBody := stdout.String()

	return 200, responseBody
}

func PostFileRequestNoAuth(url string, fileName string, folderPath string) (int, string) {
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, errFile1 := os.Open(folderPath + "/" + fileName)
	if errFile1 != nil {
		fmt.Println(errFile1)
		return 0, errFile1.Error()
	}
	defer file.Close()
	part1, _ := writer.CreateFormFile("file", filepath.Base(folderPath+"/"+fileName))
	_, err := io.Copy(part1, file)
	if err != nil {
		fmt.Println(err)
		return 0, err.Error()
	}

	err = writer.Close()
	if err != nil {
		fmt.Println(err)
		return 0, err.Error()
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return 0, err.Error()
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return 0, err.Error()
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return 0, err.Error()
	}

	return res.StatusCode, string(body)
}

func PostRequestWithJson(url string, jwt string, json string) (int, string) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(json)))
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}

func PostRequestWithJsonObject(url string, jwt string, bodyPayload interface{}) (int, string) {
	jsonPayloadBytes, err := json.Marshal(bodyPayload)
	if err != nil {
		return 0, err.Error()
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonPayloadBytes)))
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}

func PostRequest(url string, jwt string) (int, string) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}

func PostRequestNoAuth(url string) (int, string) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return 0, err.Error()
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}

func GetRequest(url string, jwt string) (int, string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}

func GetRequestNoAuth(url string) (int, string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err.Error()
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}

func GetRequestWithJSON(url string, jwt string, json string) (int, string) {
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte(json)))
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}

func PutRequestWithJson(url string, jwt string, json string) (int, string) {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer([]byte(json)))
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}

func PutRequest(url string, jwt string) (int, string) {
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}

func RestCall(url string, jwt string, callType string) (int, string) {
	req, err := http.NewRequest(callType, url, nil)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}

func RestCallNoAuth(url string, callType string) (int, string) {
	req, err := http.NewRequest(callType, url, nil)
	if err != nil {
		return 0, err.Error()
	}
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}

func DeleteRequest(url string, jwt string) (int, string) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}

func DeleteRequestWithJson(url string, jwt string, json string) (int, string) {
	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer([]byte(json)))
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err.Error()
	}

	return resp.StatusCode, string(responseBody)
}
