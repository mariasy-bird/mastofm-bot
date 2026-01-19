package state

import (
	"encoding/json"
	"log"
	"os"
)

// Last timestamp, for persistence
type LastUTS struct {
	LastUTS string `json:"last_uts"`
}

// Loading/writing the persistence file (for idempotency)
func Load(filename string) (LastUTS, error) {
	var lastuts LastUTS

	file, err := os.Open(filename)
	if err != nil {
		return LastUTS{}, err
	}
	defer file.Close() //nolint:errcheck

	err = json.NewDecoder(file).Decode(&lastuts)
	if err != nil {
		return LastUTS{}, err
	}

	return lastuts, err
}

func Save(filename string, l LastUTS) {
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Error when saving persist: %v", err)
		return
	}
	defer file.Close() //nolint:errcheck

	err = json.NewEncoder(file).Encode(l)
	if err != nil {
		log.Printf("Error when encoding persist file: %v", err)
	}
}
