package service

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ternary-club/backend/utils"
)

// Binary commands
type Commands struct {
	Start string
	End   *string
}

// Individual binary configs
type BinaryConfig struct {
	Commands   Commands `json:"commands"`
	Repository string   `json:"repository"`
	WipeFiles  bool     `json:"wipe_files"`
}

// Multiple binaries configs
type BinariesConfig map[string]BinaryConfig

// Multiple binaries locks
type BinariesLock map[string]int64

// Portfolio struct
type Portfolio struct {
	Config BinariesConfig
	Lock   BinariesLock
}

// Get configuration from file
func getBinariesConfigs() (configs BinariesConfig) {
	// Open JSON file
	file, err := os.Open("./binary.json")
	if err != nil {
		log.Fatalln("couldn't open binary.json:", err)
	}

	// Read file
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalln("couldn't read binary.json bytes:", err)
	}

	// Close reader
	file.Close()

	// Unmarshal configs
	err = json.Unmarshal(bytes, &configs)
	if err != nil {
		log.Fatalln("couldn't unmarshal binary.json into a valid map of binaries configurations:", err)
	}
	return
}

// Get versioning file
func openBinariesLockFile() (file *os.File) {
	// Check if binary-lock.json exists
	if exists := utils.Exists("./binary-lock.json"); !exists {
		return nil
	}

	// Open JSON file
	file, err := os.OpenFile("./binary-lock.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln("couldn't open binary-lock.json:", err)
	}

	return
}

// Remove old and add new binaries to binary-lock.json
func TidyEnvironment() (portfolio Portfolio) {
	// Get binary.json content
	portfolio.Config = getBinariesConfigs()
	portfolio.Lock = BinariesLock{}

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
			log.Fatalln("couldn't create binary-lock.json:", err)
		}
	} else {
		// Read binary-lock.json
		bytes, err := ioutil.ReadAll(lockFile)
		if err != nil {
			log.Fatalln("couldn't read binary-lock.json bytes:", err)
		}

		// Unmarshal binary-lock.json content
		err = json.Unmarshal(bytes, &portfolio.Lock)
		if err != nil {
			log.Fatalln("couldn't unmarshal binary-lock.json into a valid map of binaries version identifiers:", err)
		}
	}

	// Iterate through configured binaries to add
	for k := range portfolio.Config {
		_, ok := portfolio.Lock[k]
		// If it doesn't find a binary in lock, add it with version 0 to the lock
		if !ok {
			portfolio.Lock[k] = 0
		}
	}

	// Iterate through locked binaries to remove
	for k := range portfolio.Lock {
		_, ok := portfolio.Config[k]
		// Check if file is in configs
		if !ok {
			// If it isn't, delete it's directory and remove it from lock
			err := os.Remove(k)
			if err != nil {
				log.Println("couldn't delete directory", k, "while cleaning binaries:", err)
			}
			delete(portfolio.Lock, k)
		}
	}

	// Rewind reader to override contents
	lockFile.Truncate(0)
	lockFile.Seek(0, 0)
	// Marshal updated lock
	bytes, _ := json.MarshalIndent(portfolio.Lock, "", "    ")
	// Write to binary-lock.json
	lockFile.Write(bytes)

	return
}

// Update binaries and lock versions
func (portfolio *Portfolio) Update() bool {
	// Get root working directory
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	// Iterate through locked binaries to update them
	for k, v := range portfolio.Lock {
		// Open working directory
		os.Chdir(exPath)
		// Get config
		config := portfolio.Config[k]
		// Get repository info
		repoInfo := utils.FetchRepoInfo(config.Repository)
		if repoInfo == nil {
			return false
		}

		// Check versions
		if v < repoInfo.ID {
			// Update version
			portfolio.Lock[k] = repoInfo.ID
			// Create new directory
			newDir := k + "-lock"
			if !config.WipeFiles && utils.Exists(k) {
				utils.CopyDirectory(k, newDir)
			} else {
				if err := utils.CreateIfNotExists(newDir, 0755); err != nil {
					log.Println("couldn't create directory of", k+":", err)
					continue
				}
			}

			// Update binaries
			utils.DownloadFromRepo(config.Repository, newDir)

			// End running service
			if config.Commands.End != nil {
				cmd := exec.Command(*config.Commands.End)
				stdout, err := cmd.Output()
				if err != nil {
					log.Println("couldn't end service", k+":", err)
				} else {
					log.Println("service", k, "ended with output:", stdout)
				}
			}

			// Delete old directory
			err := utils.Delete(k)
			if err != nil {
				log.Println("couldn't delete previous directory of", k+":", err)
				log.Println("trying to reinitialize", k+"...")
			} else {
				err = os.Rename(newDir, k)
				if err != nil {
					log.Println("couldn't rename", newDir, "to", k)
					continue
				}
			}

			// Open dir
			os.Chdir(k)
			_, err = os.Getwd()
			if err != nil {
				log.Println("couldn't open directory", k+":", err)
				continue
			}

			// Start service
			cmd := exec.Command(config.Commands.Start)
			stdout, err := cmd.Output()
			if err != nil {
				log.Println("couldn't start service", k+":", err)
			} else {
				log.Println("service", k, "started with output:", stdout)
			}
		}
	}

	// Open working directory
	os.Chdir(exPath)
	// Marshal updated lock
	bytes, _ := json.MarshalIndent(portfolio.Lock, "", "    ")
	// Open binary-lock.json
	lockFile := openBinariesLockFile()
	// Rewind reader to override contents
	lockFile.Truncate(0)
	lockFile.Seek(0, 0)
	// Write to binary-lock.json
	_, err = lockFile.Write(bytes)
	// Close binary-lock.json reader when finished
	lockFile.Close()

	return true
}
