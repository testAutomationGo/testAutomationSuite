package localapiserver

import (
	"log"
	"net/http"
	"testAutomationSuiteGO/localApi/localApiInternal/localApiRoutes"
)

func Start() {
	r := localApiRoutes.SetupRoutes()
	log.Println("Starting local API server on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start local API server: %v", err)
	}
}
