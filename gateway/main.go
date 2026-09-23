package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type SearchRequest struct {
	QueryText   string    `json:"query_text"`
	QueryVector []float32 `json:"query_vector"`
	TopK        int       `json:"top_k"`
}

type SearchResult struct {
	ID    string  `json:"id"`
	Text  string  `json:"text"`
	Score float32 `json:"score"`
}

type RAGResponse struct {
	ContextChunks   []SearchResult `json:"context_chunks"`
	GeneratedPrompt string         `json:"generated_prompt"`
}

// Concurrently queries a worker shard using Go routines and channels
func queryWorker(nodeURL string, req SearchRequest, wg *sync.WaitGroup, resultsChan chan<- []SearchResult) {
	defer wg.Done()

	body, _ := json.Marshal(req)
	client := http.Client{Timeout: 2 * time.Second}

	resp, err := client.Post(nodeURL+"/shard/search", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Warning: Failed to reach worker %s: %v\n", nodeURL, err)
		return
	}
	defer resp.Body.Close()

	var results []SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return
	}

	resultsChan <- results
}

func handleRAGSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Distributed cluster worker shards
	workers := []string{"http://localhost:9001"}

	var wg sync.WaitGroup
	resultsChan := make(chan []SearchResult, len(workers))

	// Fan-out query concurrently across worker nodes
	for _, worker := range workers {
		wg.Add(1)
		go queryWorker(worker, req, &wg, resultsChan)
	}

	wg.Wait()
	close(resultsChan)

	// Aggregate and rank results from shards
	var allChunks []SearchResult
	for res := range resultsChan {
		allChunks = append(allChunks, res...)
	}

	// Construct the RAG Prompt for the LLM layer
	promptContext := "Context Information:\n"
	for _, chunk := range allChunks {
		promptContext += fmt.Sprintf("- [ID: %s, Score: %.2f] %s\n", chunk.ID, chunk.Score, chunk.Text)
	}
	promptContext += fmt.Sprintf("\nUser Query: %s\nAnswer:", req.QueryText)

	ragResp := RAGResponse{
		ContextChunks:   allChunks,
		GeneratedPrompt: promptContext,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ragResp)
}

func main() {
	http.HandleFunc("/rag/search", handleRAGSearch)
	fmt.Println("VectorMesh Go Gateway running on port 8080...")
	http.ListenAndServe(":8080", nil)
}
