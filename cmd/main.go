package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/saasuke-labs/kotomi/pkg/comments"
)

var commentStore = comments.NewSitePagesIndex()

// /api/site/:site-id/page/:page-id/comments
func postCommentsHandler(w http.ResponseWriter, r *http.Request) {
	vars, err := getUrlParams(r)

	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
	}

	siteId := vars["siteId"]
	pageId := vars["pageId"]

	// Decode body as a Comment
	var comment comments.Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	comment.ID = uuid.NewString()

	commentStore.AddPageComment(siteId, pageId, comment)

	json.NewEncoder(w).Encode(comment)
}

// Expecting  /api/site/:site-id/page/:page-id/comments
func getUrlParams(r *http.Request) (map[string]string, error) {
	// Parse the path manually
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	// Expected: ["api", "site", "{siteId}", "page", "{pageId}", "comments"]
	if len(parts) != 6 || parts[0] != "api" || parts[1] != "site" || parts[3] != "page" || parts[5] != "comments" {
		return nil, fmt.Errorf("invalid path")
	}

	siteId := parts[2]
	pageId := parts[4]

	return map[string]string{
		"siteId": siteId,
		"pageId": pageId,
	}, nil

}

func getCommentsHandler(w http.ResponseWriter, r *http.Request) {

	vars, err := getUrlParams(r)

	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
	}

	siteId := vars["siteId"]
	pageId := vars["pageId"]
	comments := commentStore.GetPageComments(siteId, pageId)

	json.NewEncoder(w).Encode(comments)
}

func writeJsonResponse(w http.ResponseWriter, data interface{}) {

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(data)

}

func getHealthz(w http.ResponseWriter, r *http.Request) {

	jsonResponse := struct {
		Message string `json:"message,omitempty"`
	}{
		Message: "OK",
	}

	writeJsonResponse(w, jsonResponse)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", getHealthz)

	// mux.HandleFunc("GET /api/site/{siteId}/page/{pageId}/comments", getCommentsHandler)
	// mux.HandleFunc("POST /api/site/{siteId}/page/{pageId}/comments", postCommentsHandler)

	log.Printf("Server running at http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
