package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattn/go-mastodon"

	"mastofm-bot/internal/lastfm"
	mastoUtil "mastofm-bot/internal/mastodon"

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

// Last timestamp, for persistence
type LastUTS struct {
	LastUTS string `json:"last_uts"`
}

// Loading/writing the persistence file (for idempotency)
func loadLastUTS(filename string) (LastUTS, error) {
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

func saveLastUTS(filename string, l LastUTS) {
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

// Loading the config file to the config struct
func loadConfig(filename string) Config {
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

// Deduplication of tracks
func isNewTrack(track *lastfm.Track, lastuts LastUTS) bool {

	if track == nil {
		return false
	}

	if track.Date.UTS == "" {
		return false
	}

	return track.Date.UTS != lastuts.LastUTS
}

func main() {
	// Graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received")
		cancel()
	}()

	// Config and dedupe location
	config := loadConfig("config.json")
	persistFile := "persist.json"

	// Mastodon client
	mastoConfig := &mastodon.Config{
		Server:       config.MastodonServer,
		ClientID:     config.MastodonClientID,
		ClientSecret: config.MastodonSecret,
		AccessToken:  config.MastodonToken,
	}
	mastoClient := mastodon.NewClient(mastoConfig)

	// Ticker for polling
	var PollRate = time.Duration(config.PollRateSeconds) * time.Second
	pollTicker := time.NewTicker(PollRate)
	defer pollTicker.Stop() //nolint:errcheck

	// Loading persist file for dedupe

	lastuts, err := loadLastUTS(persistFile)
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Error when loading persist file: %v", err)
	}

	// Main loop

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down cleanly.")
			return
		case <-pollTicker.C:
			// Retrieve newest track
			track, err := lastfm.GetRecentTrack(ctx, config.LfmUsername, config.LfmApiKey)
			if err != nil {
				log.Printf("Error getting most recent track: %v", err)
			}
			// If track is new
			if track != nil && isNewTrack(track, lastuts) {
				log.Printf("New track: %s - %s\n", track.Artist.Text, track.Name)
				// Posting to Mastodon
				if !(config.TestMode) {
					mastoPost := mastodon.Toot{Status: mastoUtil.FormatPost(track)}
					toot, err := mastoClient.PostStatus(ctx, &mastoPost)
					if err != nil {
						log.Printf("Error posting to Mastodon: %#v\n", err)
					}
					log.Println("Posted: ", toot.Content)
				}
				// Saving persistence data
				lastuts.LastUTS = track.Date.UTS
				saveLastUTS(persistFile, lastuts)
			}
		}
	}
}
