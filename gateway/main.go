package main

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SearchRequest represents the incoming user query and vector
type SearchRequest struct {
	QueryText   string    `json:"query_text"`
	QueryVector []float32 `json:"query_vector"`
	TopK        int       `json:"top_k"`
}

// SearchResult represents a retrieved document chunk
type SearchResult struct {
	ID    string  `json:"id"`
	Text  string  `json:"text"`
	Score float32 `json:"score"`
}

// RAGResponse represents the final assembled prompt and chunks
type RAGResponse struct {
	ContextChunks   []SearchResult `json:"context_chunks"`
	GeneratedPrompt string         `json:"generated_prompt"`
	Cached          bool           `json:"cached"`
}

// --- Thread-Safe LRU Cache Implementation ---
type cacheItem struct {
	key   string
	value RAGResponse
}

type LRUCache struct {
	capacity  int
	evictList *list.List
	items     map[string]*list.Element
	lock      sync.Mutex
}

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity:  capacity,
		evictList: list.New(),
		items:     make(map[string]*list.Element),
	}
}

func (c *LRUCache) Get(key string) (RAGResponse, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if elem, exists := c.items[key]; exists {
		c.evictList.MoveToFront(elem)
		return elem.Value.(*cacheItem).value, true
	}
	return RAGResponse{}, false
}

func (c *LRUCache) Put(key string, value RAGResponse) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Check if already exists
	if elem, exists := c.items[key]; exists {
		c.evictList.MoveToFront(elem)
		elem.Value.(*cacheItem).value = value
		return
	}

	// Evict oldest if at capacity
	if c.evictList.Len() >= c.capacity {
		oldest := c.evictList.Back()
		if oldest != nil {
			c.evictList.Remove(oldest)
			kv := oldest.Value.(*cacheItem)
			delete(c.items, kv.key)
		}
	}

	// Add new item
	item := &cacheItem{key: key, value: value}
	elem := c.evictList.PushFront(item)
	c.items[key] = elem
}

// Initialize a global cache with a capacity of 100 items
var queryCache = NewLRUCache(100)

// Helper to generate a cache key from the query text
func getCacheKey(req SearchRequest) string {
	return req.QueryText
}

// Concurrently queries a worker shard using Go routines
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

	cacheKey := getCacheKey(req)

	// 1. Check LRU Cache first (Cache Hit)
	if cachedResp, found := queryCache.Get(cacheKey); found {
		cachedResp.Cached = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cachedResp)
		return
	}

	// 2. Cache Miss: Perform Distributed Fan-Out
	workers := []string{"http://localhost:9001"}

	var wg sync.WaitGroup
	resultsChan := make(chan []SearchResult, len(workers))

	for _, worker := range workers {
		wg.Add(1)
		go queryWorker(worker, req, &wg, resultsChan)
	}

	wg.Wait()
	close(resultsChan)

	var allChunks []SearchResult
	for res := range resultsChan {
		allChunks = append(allChunks, res...)
	}

	// Construct RAG Prompt
	promptContext := "Context Information:\n"
	for _, chunk := range allChunks {
		promptContext += fmt.Sprintf("- [ID: %s, Score: %.2f] %s\n", chunk.ID, chunk.Score, chunk.Text)
	}
	promptContext += fmt.Sprintf("\nUser Query: %s\nAnswer:", req.QueryText)

	ragResp := RAGResponse{
		ContextChunks:   allChunks,
		GeneratedPrompt: promptContext,
		Cached:          false,
	}

	// 3. Store in LRU Cache for future requests
	queryCache.Put(cacheKey, ragResp)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ragResp)
}

func main() {
	http.HandleFunc("/rag/search", handleRAGSearch)
	fmt.Println("VectorMesh Go Gateway with LRU Cache running on port 8080...")
	http.ListenAndServe(":8080", nil)
}
