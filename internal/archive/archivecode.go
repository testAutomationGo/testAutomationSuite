package archive

/*

const (
	MAX_SIZE_REGULAR_UPLOAD = 100 * 1024 * 1024 // Uploead threshold
	CHUNK_SIZE              = 10 * 1024 * 1024  // Chunk size
)

//the below code is from the
if stats.Size() > MAX_SIZE_REGULAR_UPLOAD {
		return uploadWithTUS(filePath, groupId, name, verbose, stats, network)
	}

func uploadWithTUS(filePath string, groupId string, name string, verbose bool, stats os.FileInfo, network string) (UploadResponse, error) {
	jwt, err := FindToken()
	if err != nil {
		return UploadResponse{}, err
	}

	// Create the TUS client with config
	config := &tus.Config{
		ChunkSize:  CHUNK_SIZE, // 50MB chunks
		Resume:     false,
		Header:     http.Header{"Authorization": {fmt.Sprintf("Bearer %s", jwt)}},
		HttpClient: http.DefaultClient,
	}

	uploadHost := GetUploadsHost()
	url := fmt.Sprintf("https://%s/v3/files", uploadHost)
	client, err := tus.NewClient(url, config)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("failed to create TUS client: %w", err)
	}

	// Open the file
	f, err := os.Open(filePath)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	networkParam, err := GetNetworkParam(network)
	if err != nil {
		return UploadResponse{}, err
	}

	// Create metadata
	metadata := map[string]string{
		"filename": filepath.Base(filePath),
		"network":  networkParam,
	}
	if groupId != "" {
		metadata["group_id"] = groupId
	}
	if name != "nil" {
		metadata["filename"] = name
	}

	// Create the upload
	upload := tus.NewUpload(f, stats.Size(), metadata, "")

	// Create and configure the uploader
	uploader, err := client.CreateUpload(upload)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("failed to create upload: %w", err)
	}

	var bar *progressbar.ProgressBar
	if verbose {
		fmt.Printf("Starting upload of %s (%s)\n", stats.Name(), formatSize(int(stats.Size())))
		bar = progressbar.NewOptions64(
			stats.Size(),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionShowBytes(true),
			progressbar.OptionSetDescription("Uploading..."),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "█",
				SaucerPadding: " ",
				BarStart:      "|",
				BarEnd:        "|",
			}),
			progressbar.OptionOnCompletion(cmpl),
		)

		go func() {
			for {
				offset := uploader.Offset()
				if offset >= stats.Size() {
					return
				}
				bar.Set64(offset)
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	err = uploader.Upload()
	if err != nil {
		return UploadResponse{}, fmt.Errorf("failed during upload: %w", err)
	}

	if verbose {
		fmt.Println("\nUpload completed!")
	}

	uploadURL := uploader.Url()
	urlParts := strings.Split(uploadURL, "/")
	fileId := urlParts[len(urlParts)-2]

	apiURL := fmt.Sprintf("https://%s/v3/files/%s", GetAPIHost(), fileId)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("failed to create response request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+string(jwt))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("failed to fetch upload response: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var response UploadResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return UploadResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}

	formattedJSON, err := json.MarshalIndent(response.Data, "", "    ")
	if err != nil {
		return UploadResponse{}, fmt.Errorf("failed to format response: %w", err)
	}
	fmt.Println(string(formattedJSON))

	return response, nil
}

func createMultipartRequest(filePath string, files []string, body io.Writer, stats os.FileInfo, groupId string, name string, network string) (string, error) {
	contentType := ""
	writer := multipart.NewWriter(body)

	fileIsASingleFile := !stats.IsDir()
	for _, f := range files {
		file, err := os.Open(f)
		fmt.Println("File Path:", f)
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
			fmt.Println("Single File Path Base:", filepath.Base(f))
		} else {
			relPath, _ := filepath.Rel(filePath, f)
			fileName := path.Join(stats.Name(), relPath)
			part, err = writer.CreateFormFile("file", fileName)
			//part, err = writer.CreateFormFile("file", filepath.Join(stats.Name(), relPath))
			fmt.Println("Folder File Path Base:", filepath.Join(stats.Name(), relPath))
		}
		if err != nil {
			return contentType, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return contentType, err
		}
	}

	networkParam, err := GetNetworkParam(network)
	if err != nil {
		return "", err
	}

	err = writer.WriteField("network", networkParam)

	if err != nil {
		return contentType, err
	}

	if groupId != "" {
		err := writer.WriteField("group_id", groupId)
		if err != nil {
			return contentType, err
		}
	}

	nameToUse := stats.Name()
	if name != "nil" {
		nameToUse = name
	}
	err = writer.WriteField("name", nameToUse)
	if err != nil {
		return contentType, err
	}

	err = writer.Close()
	if err != nil {
		return contentType, err
	}

	contentType = writer.FormDataContentType()

	return contentType, nil
}
*/
