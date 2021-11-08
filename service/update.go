package service

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/ternary-club/backend/utils"
)

const TERNARY_CLUB_ORG = "ternary-club"

var REPOS = [...]string{"versioning-test", "terry"}

// Versions of binaries
var Versions = map[string]uint64{}

// Get versioning file
func openVersionsFile() (file *os.File) {
	// Check if lock.json exists
	if exists := utils.Exists("./lock.json"); !exists {
		return nil
	}

	// Open JSON file
	file, err := os.OpenFile("./lock.json", os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatalln("couldn't open lock.json:", err)
	}

	return
}

// Remove old and add new binaries to lock.json
func Update() {
	// Open lock.json
	versionsFile := openVersionsFile()
	// Close lock.json reader when finished
	defer versionsFile.Close()

	// Create lock.json if doesn't exist
	if versionsFile == nil {
		var err error
		// Create lock.json
		versionsFile, err = os.Create("./lock.json")
		if err != nil {
			log.Fatalln("couldn't create lock.json:", err)
		}
	} else {
		// Read lock.json
		bytes, err := ioutil.ReadAll(versionsFile)
		if err != nil {
			log.Fatalln("couldn't read lock.json bytes:", err)
		}

		// Unmarshal lock.json content
		err = json.Unmarshal(bytes, &Versions)
		if err != nil {
			log.Fatalln("couldn't unmarshal lock.json into a valid map of binaries IDs:", err)
		}
	}

	// Update binaries
	for _, r := range REPOS {
		v, ok := Versions[r]
		if id := utils.FetchRepoInfo(TERNARY_CLUB_ORG + "/" + r).ID; !ok || id != v {
			update(r)
			Versions[r] = id
		}
	}

	// Rewind pointer to override previous bytes
	versionsFile.Truncate(0)
	versionsFile.Seek(0, 0)
	// Marshal updated lock
	bytes, _ := json.Marshal(Versions)
	// Write to binary-lock.json
	versionsFile.Write(bytes)
}

// Update binaries and lock versions
func update(repo string) {
	/// Delete old directory
	if utils.Exists(repo) {
		utils.Delete(repo)
	}
	// Create new directory
	if err := os.MkdirAll("bin/"+repo, 0755); err != nil {
		log.Fatalln("couldn't create directory: \""+repo+"\":", err)
	}
	// Update binaries
	utils.DownloadFromRepo(TERNARY_CLUB_ORG+"/"+repo, "bin/"+repo)
}
