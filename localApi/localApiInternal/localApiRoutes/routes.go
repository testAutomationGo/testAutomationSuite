package localApiRoutes

import (
	"testAutomationSuiteGO/localApi/localApiInternal/localApiHandlers"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/hello", localApiHandlers.HelloHandler).Methods("GET")
	return r
}
