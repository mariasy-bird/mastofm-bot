package config

import (
	"encoding/json"
	"log"
	"os"
)

// Configuration data
type Config struct {
	MastodonServer   string `json:"mastodon_server"`
	MastodonClientID string `json:"mastodon_client_id"`
	MastodonSecret   string `json:"mastodon_secret"`
	MastodonToken    string `json:"mastodon_token"`
	LfmUsername      string `json:"lfm_username"`
	LfmApiKey        string `json:"lfm_api_key"`
	PollRateSeconds  int    `json:"poll_rate"`
	TestMode         bool   `json:"test_mode"`
}

// Loading the config file to the config struct
func Load(filename string) Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Could not open config file: %v", err)
	}
	defer file.Close() //nolint:errcheck

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		log.Fatalf("Cannot parse config JSON: %v", err)
	}

	return cfg
}
