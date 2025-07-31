package apiFunctions

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func FolderUpload(url, jwt, folderPath, name string) (int, string) {

	stats, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		return 0, "Dir does not exist."
	}
	fmt.Println("Dir does exist.")
	files, err := PathsFinder(folderPath, stats)
	if err != nil {
		return 0, err.Error()
	}

	body := &bytes.Buffer{}
	contentType, err := CreateMultipartRequest(folderPath, files, body, stats, name)
	if err != nil {
		return 0, err.Error()
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return 0, err.Error()
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("content-type", contentType)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}

	defer res.Body.Close()

	resbody, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err.Error()
	}

	return res.StatusCode, string(resbody)
}

func CreateMultipartRequest(filePath string, files []string, body io.Writer, stats os.FileInfo, name string) (string, error) {
	contentType := ""
	writer := multipart.NewWriter(body)

	fileIsASingleFile := !stats.IsDir()
	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			return contentType, err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				log.Fatal("could not close file")
			}
		}(file)

		var part io.Writer
		if fileIsASingleFile {
			part, err = writer.CreateFormFile("file", filepath.Base(f))
		} else {
			relPath, _ := filepath.Rel(filePath, f)
			if runtime.GOOS == "windows" {
				relPathForward := strings.ReplaceAll(relPath, "\\", "/")
				folderName := stats.Name()
				folderNameForward := strings.ReplaceAll(folderName, "\\", "/")
				fullPath := folderNameForward
				if relPathForward != "" {
					fullPath = folderNameForward + "/" + relPathForward
				}
				part, err = writer.CreateFormFile("file", fullPath)
			} else {
				part, err = writer.CreateFormFile("file", filepath.Join(stats.Name(), relPath))
			}
		}
		if err != nil {
			return contentType, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return contentType, err
		}
	}

	err := writer.Close()
	if err != nil {
		return contentType, err
	}

	contentType = writer.FormDataContentType()

	return contentType, nil
}

func PathsFinder(filePath string, stats os.FileInfo) ([]string, error) {
	var err error
	files := make([]string, 0)
	fileIsASingleFile := !stats.IsDir()
	if fileIsASingleFile {
		files = append(files, filePath)
		return files, err
	}
	err = filepath.Walk(filePath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})

	if err != nil {
		return nil, err
	}

	return files, err
}

type PinataOptions struct {
	CidVersion int    `json:"cidVersion"`
	GroupId    string `json:"groupId,omitempty"`
}

type PinataMetadata struct {
	Name      string            `json:"name"`
	KeyValues map[string]string `json:"keyvalues,omitempty"`
}
