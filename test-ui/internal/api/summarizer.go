// Route handler to expose summary endpoint via localhost
package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

type SummaryRequest struct {
	Text string `json:"text"`
}

type SummaryResponse struct {
	Summary string `json:"summary"`
}

func summarizeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SummaryRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	summary, err := SummarizeWithOllama(req.Text)
	if err != nil {
		http.Error(w, "Failed to summarize", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(SummaryResponse{Summary: summary})
}

func main() {
	// Register the summary route
	http.HandleFunc("/api/summarize", summarizeHandler)
	log.Println("Summary endpoint running at http://localhost:8081/api/summarize")
	http.ListenAndServe(":8081", nil)
}

