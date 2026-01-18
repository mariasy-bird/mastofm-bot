package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mattn/go-mastodon"
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

// Track data
type lfmTrack struct {
	Name   string `json:"name"`
	Artist struct {
		Text string `json:"#text"`
	} `json:"artist"`
	Album struct {
		Text string `json:"#text"`
	} `json:"album"`
	Date struct {
		UTS string `json:"uts"`
	} `json:"date"`
}

// Last timestamp, for persistence
type LastUTS struct {
	LastUTS string `json:"last_uts"`
}

// Fetch all recent tracks from Last.fm, return an lfmTrack struct
func lfmGetRecentTrack(lfmUsername, lfmApiKey string) (*lfmTrack, error) {
	url := "https://ws.audioscrobbler.com/2.0" +
		"?method=user.getRecentTracks" +
		"&user=" + lfmUsername +
		"&api_key=" + lfmApiKey +
		"&format=json&limit=1"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		RecentTracks struct {
			Track []lfmTrack `json:"track"`
		} `json:"recenttracks"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.RecentTracks.Track) == 0 {
		return nil, fmt.Errorf("last.fm returned no tracks")
	}
	return &parsed.RecentTracks.Track[0], nil
}

// Loading/writing the persistence file (for idempotency)
func loadLastUTS(filename string) (LastUTS, error) {
	var lastuts LastUTS

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Could not open persist file: %v", err)
		return LastUTS{}, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&lastuts)
	if err != nil {
		log.Printf("Error: Unable to decode persist file: %v", err)
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
	defer file.Close()

	json.NewEncoder(file).Encode(l)
}

// Loading the config file to the config struct
func loadConfig(filename string) Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Could not open config file: %v", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		log.Fatalf("Cannot parse config JSON: %v", err)
	}

	return cfg
}

// Deduplication of tracks
func isNewTrack(track *lfmTrack, lastuts LastUTS) bool {
	if track.Date.UTS == "" {
		return false
	}
	if track == nil {
		return false
	}

	return track.Date.UTS != lastuts.LastUTS
}

// Mastodon post format

func formatPost(track *lfmTrack) mastodon.Toot {

	if track.Album.Text != "" {
		post := "🎵 Now listening\n" + 
			track.Artist.Text + " - " + track.Name +
			"\n 📀 " + track.Album.Text
		toot := mastodon.Toot{
			Status: post,
		}
		return toot
	}
	// Sometimes album is missing.
	post := "🎵 Now listening\n" + 
		track.Artist.Text + " - " + track.Name +
		"\n 📀 Unknown Album"
	toot := mastodon.Toot{
		Status: post,
	}
	return toot
}

func main() {
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
	defer pollTicker.Stop()

	// Loading persist file for dedupe

	lastuts, err := loadLastUTS(persistFile)
	if err != nil {
		log.Printf("Error when loading persist file: %v", err)
	}

	// Main loop

	for range pollTicker.C {
		// Retrieve newest track
		track, err := lfmGetRecentTrack(config.LfmUsername, config.LfmApiKey)
		if err != nil {
			log.Printf("Error getting most recent track: %v", err)
		}
		// If track is new
		if track != nil && isNewTrack(track, lastuts) {
			log.Printf("New track: %s - %s\n", track.Artist.Text, track.Name)
			// Posting to Mastodon
			mastoPost := formatPost(track)
			toot, err := mastoClient.PostStatus(context.Background(), &mastoPost)
			if err != nil {
				log.Printf("Error posting to Mastodon: %#v\n", err)
			}
			log.Println("Posted: ", toot.Content)
			// Saving persistence data
			lastuts.LastUTS = track.Date.UTS
			saveLastUTS(persistFile, lastuts)
		}
	}
}
