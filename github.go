package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
)

// Based on https://gist.github.com/metal3d/002e4f0d8545f83c2ace

// Prepare GET request
func makeRequest(url string) *http.Request {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "metal3d-go-client")
	return req
}

// Download resource from given URL and write 1 in chan when finished
func downloadResource(id float64, c chan int) {
	defer func(){ c <- 1 }()
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/assets/%.0f", repo, id)
	fmt.Printf("Start: %s\n", url)
	req := makeRequest(url)

	req.Header.Add("Accept", "application/octet-stream")

	client := http.Client{}
	resp, _ := client.Do(req)

	disp := resp.Header.Get("Content-disposition")
	re := regexp.MustCompile(`filename=(.+)`)
	matches := re.FindAllStringSubmatch(disp, -1)

	if len(matches) == 0 || len(matches[0]) == 0 {
		log.Println("WTF: ", matches)
		log.Println(resp.Header)
		log.Println(req)
		return
	}

	disp = matches[0][1]

	f, err := os.OpenFile(disp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		log.Fatal(err)
	}

	b := make([]byte, 4096)
	var i int

	for err == nil {
		i, err = resp.Body.Read(b)
		f.Write(b[:i])
	}
	fmt.Printf("Finished: %s -> %s\n", url, disp)
	f.Close()
}
