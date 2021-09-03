package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

// Download a file from a URL
func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// Download latest version of the repository
func DownloadLatest(config BinaryConfig) bool {
	// Build URL
	url := fmt.Sprintf("https://api.github.com/repos/" + config.Repository + "/releases/latest")
	// Send request
	req, _ := http.NewRequest("GET", url, nil)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("couldn't get response from", url+":", err)
	}
	// Close body reader when it finishes
	defer resp.Body.Close()

	// Read body if successful
	if resp.StatusCode != http.StatusOK {
		log.Println("couldn't fetch", url+": status code", resp.StatusCode)
		return false
	}

	// Body structs
	type asset struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}
	type body struct {
		ID     int64   `json:"id"`
		Assets []asset `json:"assets"`
	}

	// Marshal body
	b := &body{}
	err = json.NewDecoder(resp.Body).Decode(&b)
	if err != nil {
		log.Println("couldn't marshal", url, "body to JSON:", err)
	}
	// Check assets
	if len(b.Assets) == 0 {
		return true
	}

	// Get repo name
	path := strings.Split(config.Repository, "/")
	dirName := path[len(path)-1]
	// Create directory if it doesn't exist
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err = os.Mkdir(dirName, 0755)
		if err != nil {
			log.Println("couldn't create directory", dirName+":", err)
		}
	}

	// Download files
	wg := sync.WaitGroup{}
	wg.Add(len(b.Assets))
	for _, v := range b.Assets {
		go func(a asset, path string, wg *sync.WaitGroup) {
			err := DownloadFile(path+"/"+a.Name, a.BrowserDownloadURL)
			if err != nil {
				log.Println("couldn't download file", a.Name+":", err)
			}
			log.Println("downloading", path+"/"+a.Name+"...")
			wg.Done()
		}(v, dirName, &wg)
	}
	wg.Wait()
	log.Println("finished updating", config.Repository+"!")

	return true
}
