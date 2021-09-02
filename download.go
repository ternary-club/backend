package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/prometheus/common/log"
)

func DownloadLatest(b BinaryConfig) {
	// Build URL
	url := fmt.Sprintf("http://https://api.github.com/repos/", b.Repository, "/releases/latest")
	// Send request
	req, _ := http.NewRequest("GET", url, nil)
	client := http.Client{}
	resp, _ := client.Do(req)
	// Close body reader when it finishes
	defer resp.Body.Close()

	// Read body if successful
	if resp.StatusCode == http.StatusOK {
		// Read bytes
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		// Marshal body
		bodyJSON, err := json.Marshal(bodyBytes)
		if err != nil {
			log.Println("couldn't marshal", url, "body to JSON:", err)
		}

	} else {
		log.Println("couldn't fetch", url+": status code", resp.StatusCode)
	}
}
