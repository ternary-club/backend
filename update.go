package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

// Check if file or directory exists
func checkExists(path string) bool {
	// Check if file exists
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("couldn't check stat of", path+":", err)
		}
		return false
	}
	return true
}

// Get version of specified binary
func getVersion(path string) (string, error) {
	if exists := checkExists(path); !exists {
		return "0", nil
	} else {
		out, err := exec.Command(path, "-v").Output()
		if err != nil {
			return "0", err
		} else {
			return string(out), nil
		}
	}
}

// Individual binary configs
type BinaryConfig struct {
	Command    string   `json:"command"`
	Repository string   `json:"repository"`
	Content    []string `json:"content"`
}

// Multiple binaries configs
type BinariesConfig map[string]BinaryConfig

// Individual binary locks
type BinaryLock struct {
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
}

// Multiple binaries locks
type BinariesLock map[string]BinaryLock

// Get configuration from file
func getBinariesConfigs() (configs BinariesConfig) {
	// Open JSON file
	file, err := os.Open("./binary.json")
	if err != nil {
		log.Fatal("couldn't open binary.json:", err)
	}
	// Read file
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal("couldn't read binary.json bytes:", err)
	}
	// Close reader
	file.Close()
	// Unmarshal configs
	err = json.Unmarshal(bytes, &configs)
	if err != nil {
		log.Fatal("couldn't unmarshal binary.json into a valid map of binaries configurations:", err)
	}
	return
}

// Normalize binary-lock.json
func TidyBinaryLock() {
	// Binaries configs
	configs := getBinariesConfigs()
	// Final binary lock
	lock := BinariesLock{}
	// Check if binary-lock.json exists
	if exists := checkExists("./binary-lock.json"); exists {
		// Open JSON file
		file, err := os.Open("./binary-lock.json")
		if err != nil {
			log.Fatal("couldn't open binary-lock.json:", err)
		}
		// Read file
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal("couldn't read binary-lock.json bytes:", err)
		}
		// Close reader
		file.Close()
		// Unmarshal configs
		err = json.Unmarshal(bytes, &lock)
		if err != nil {
			log.Fatal("couldn't unmarshal binary-lock.json into a valid map of binaries contents lock:", err)
		}
	}
	// Iterate through expected binaries
	for k, v := range configs {
		_, ok := lock[k]
		// If it doesn't find a binary, add it with version 0 to the lock
		if !ok {
			lock[k] = BinaryLock{Version: "0"}
		}
	}
}

// Get configuration from file
func GetBinariesLock() (lock BinariesLock) {
	// Check if binary-lock.json exists
	if exists := checkExists("./binary-lock.json"); !exists {
		return nil
	}
	// Open JSON file
	file, err := os.Open("./binary-lock.json")
	if err != nil {
		log.Fatal("couldn't open binary-lock.json:", err)
	}
	// Read file
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal("couldn't read binary-lock.json bytes:", err)
	}
	// Close reader
	file.Close()
	// Unmarshal configs
	err = json.Unmarshal(bytes, &lock)
	if err != nil {
		log.Fatal("couldn't unmarshal binary-lock.json into a valid map of binaries contents lock:", err)
	}
	return
}

// Download single binary contents
