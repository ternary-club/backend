package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

// Individual binary configs
type BinaryConfig struct {
	Command    string `json:"command"`
	Repository string `json:"repository"`
	WipeFiles  bool   `json:"wipeFiles"`
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

// Get configuration from file
func openBinariesLockFile() (file *os.File) {
	// Check if binary-lock.json exists
	if exists := checkExists("./binary-lock.json"); !exists {
		return nil
	}
	// Open JSON file
	file, err := os.Open("./binary-lock.json")
	if err != nil {
		log.Fatalln("couldn't open binary-lock.json:", err)
	}
	return
}

// Remove old and add new binaries to binary-lock.json
func TidyEnvironment() (configs BinariesConfig, lock BinariesLock) {
	// Get binary.json content
	configs = getBinariesConfigs()
	// Open binary-lock.json
	lockFile := openBinariesLockFile()
	// Close binary-lock.json reader when finished
	defer lockFile.Close()
	// Create binary-lock.json if doesn't exist
	if lockFile == nil {
		var err error
		// Create binary-lock.json
		lockFile, err = os.Create("./binary-lock.json")
		if err != nil {
			log.Fatal("couldn't create binary-lock.json:", err)
		}
	} else {
		// Read binary-lock.json
		bytes, err := ioutil.ReadAll(lockFile)
		if err != nil {
			log.Fatal("couldn't read binary-lock.json bytes:", err)
		}
		// Unmarshal binary-lock.json content
		err = json.Unmarshal(bytes, &lock)
		if err != nil {
			log.Fatal("couldn't unmarshal binary-lock.json into a valid map of binaries version identifiers:", err)
		}
	}

	// Iterate through configured binaries to add
	for k := range configs {
		_, ok := lock[k]
		// If it doesn't find a binary in lock, add it with version 0 to the lock
		if !ok {
			lock[k] = BinaryLock{Version: "0"}
		}
	}
	// Iterate through locked binaries to remove
	for k := range lock {
		_, ok := configs[k]
		// Check if file is in configs
		if !ok {
			// If it isn't, delete it's directory and remove it from lock
			err := os.Remove(k)
			if err != nil {
				fmt.Println("couldn't delete directory", k, "while cleaning binaries:", err)
			}
			delete(lock, k)
		}
	}

	// Rewind reader to override contents
	lockFile.Truncate(0)
	lockFile.Seek(0, 0)
	// Marshal updated lock
	bytes, _ := json.Marshal(lock)
	// Write to binary-lock.json
	lockFile.Write(bytes)

	return
}
