package state

import (
	"encoding/json"
	"log"
	"os"
)

type Persistent struct {
	LastUTS    int64  `json:"last_uts"`
	Name       string `json:"last_track_name"`
	AlbumName  string `json:"last_album_name"`
	ArtistName string `json:"last_artist_name"`
}

// Loading/writing the persistence file (for idempotency)
func Load(filename string) (Persistent, error) {
	var lastuts Persistent

	file, err := os.Open(filename)
	if err != nil {
		return Persistent{}, err
	}
	defer file.Close() //nolint:errcheck

	err = json.NewDecoder(file).Decode(&lastuts)
	if err != nil {
		return Persistent{}, err
	}

	return lastuts, err
}

func Save(filename string, l Persistent) {
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
