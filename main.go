package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// PageRequest represents a request to download the source of a webpage.
type PageRequest struct {
	URI        string `json:"uri"`
	RetryLimit int    `json:"retryLimit"`
}

// PageResponse represents the response to a PageRequest.
type PageResponse struct {
	ID        string `json:"id"`
	URI       string `json:"uri"`
	SourceURI string `json:"sourceUri"`
}

const cacheExpiryDuration = 24 * time.Hour

var cache = make(map[string]*CachedPage)

// CachedPage represents a webpage that has been downloaded and cached.
type CachedPage struct {
	ID        string
	URI       string
	SourceURI string
	Timestamp time.Time
}

func main() {
	fmt.Println("Running server at 7771 port......")
	http.HandleFunc("/pagesource", handlePageSource)
	http.ListenAndServe(":7771", nil)
}

func handlePageSource(w http.ResponseWriter, r *http.Request) {
	var req PageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check the cache to see if the page has already been requested within the past 24 hours.
	if page, ok := cache[req.URI]; ok {
		if time.Since(page.Timestamp) < cacheExpiryDuration {
			// Serve the page from the cache.
			json.NewEncoder(w).Encode(page)
			return
		}
	}

	// Download the webpage and cache it.
	id, sourceURI, err := downloadPage(req.URI, req.RetryLimit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("helllo, making food ready to send.....")
	fmt.Println(id)
	fmt.Println(sourceURI)
	fmt.Printf("sent.....\n---------")

	page := &CachedPage{
		ID:        id,
		URI:       req.URI,
		SourceURI: sourceURI,
		Timestamp: time.Now(),
	}
	cache[req.URI] = page

	json.NewEncoder(w).Encode(page)
}

func downloadPage(uri string, retryLimit int) (string, string, error) {
	id := generateID()
	sourceURI := fmt.Sprintf("/files/%s.html", id)
	fmt.Println("-------\nHelloooo,test")
	fmt.Println(sourceURI)
	fmt.Println(id)

	// Set the retry limit to the minimum of the specified value and 10.
	if retryLimit > 10 {
		retryLimit = 10
	}

	var err error
	for i := 0; i < retryLimit; i++ {
		resp, err := http.Get(uri)
		if err == nil {
			defer resp.Body.Close()
			page, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				// Write the webpage to the local file system.
				if err := ioutil.WriteFile(sourceURI, page, 0644); err != nil {
					return "", "", err
				}
				fmt.Printf(id)
				fmt.Printf(sourceURI)

				return id, sourceURI, nil
			}
		}
	}
	return "", "", err
}

func generateID() string {
	// This function generates a unique ID.
	fmt.Printf("strinf returns: ")
	return fmt.Sprintf("1672250592852871000")
}
