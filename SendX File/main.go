package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
)

// Initialize the cache with a maximum size of 100KB and an eviction function
var cache = make(map[string][]byte, 1000000)
var cacheSize = 1000000
var cacheSizeUsed = 0

func evict(url string, data []byte) {
	delete(cache, url)
	cacheSizeUsed -= len(data)
}

func main() {
	fmt.Println("Running at https://localhost:8080")
	// fmt.Println("SendX Assignment| Mrityunjaya Tiwari")
	http.HandleFunc("/", downloadHandler)
	http.ListenAndServe(":8080", nil)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Define the cache data structure and eviction policy

	if r.Method == "POST" {
		// Read the URL from the form
		url := r.FormValue("url")

		// Check if the file data is already present in the cache
		data, ok := cache[url]
		if ok {
			// File data is present in the cache, so write it to the response
			w.Header().Set("X-File-Source", "cache")
			writeFileDataToResponse(w, url, data, nil, nil)
			return
		}

		// File data is not present in the cache, so send an HTTP GET request to the URL
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Error retrieving file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Check the HTTP status code
		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Error retrieving file: "+resp.Status, http.StatusInternalServerError)
			return
		}

		// Read the file data from the HTTP response
		data, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading file data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Store the file data in the cache
		cache[url] = data
		cacheSizeUsed += len(data)

		// Evict items from the cache if necessary to keep the cache size below the maximum
		for cacheSizeUsed > cacheSize {
			for url, data := range cache {
				evict(url, data)
				break
			}
		}

		// Write the file data to the response
		w.Header().Set("X-File-Source", "url")
		writeFileDataToResponse(w, url, data, resp, err)
		return
	}

	// Render the form template
	tmpl, err := template.ParseFiles("form.html")
	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func writeFileDataToResponse(w http.ResponseWriter, url string, data []byte, resp *http.Response, err error) {
	// Read the content disposition header to get the file name
	var fileName string
	if resp != nil {
		contentDisposition := resp.Header.Get("Content-Disposition")
		if contentDisposition != "" {
			_, params, err := mime.ParseMediaType(contentDisposition)
			if err == nil {
				fileName = params["filename"]
			}
		}
	}
	if fileName == "" {
		// Use the last part of the URL as the file name if no content disposition header is present
		parts := strings.Split(url, "/")
		fileName = parts[len(parts)-1]
	}

	// Set the headers for the response
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)

	// Write the file data to the response
	if err == nil {
		_, err = w.Write(data)
	}
	if err != nil {
		http.Error(w, "Error writing to response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
