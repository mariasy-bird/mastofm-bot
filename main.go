package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattn/go-mastodon"

	"mastofm-bot/internal/config"
	"mastofm-bot/internal/lastfm"
	mastoUtil "mastofm-bot/internal/mastodon"
	"mastofm-bot/internal/state"
)

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
	configPath := flag.String(
		"config",
		"config.json",
		"Path to config file",
	)
	statePath := flag.String(
		"state",
		"state.json",
		"Path to persistent state file",
	)
	flag.Parse()

	if _, err := os.Stat(*configPath); err != nil {
		log.Fatalf("Config file not found: %s", *configPath)
	}

	config := config.Load(*configPath)
	persistFile := *statePath

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

	lastuts, err := state.Load(persistFile)
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
			if track != nil && lastfm.IsNew(track, lastuts) {
				log.Printf("New track: %s - %s\n", track.Artist.Text, track.Name)
				var mediaIDs []mastodon.ID
				// Get best track art
				imgURL := track.BestImageURL()
				// Download art
				if imgURL != "" && config.AlbumArt {
					imgBytes, err := mastoUtil.DownloadImage(ctx, imgURL)
					if err == nil {
						media, err := mastoUtil.UploadAlbumArt(ctx, mastoClient, imgBytes, track)
						if err == nil {
							mediaIDs = []mastodon.ID{media.ID}
						}
					}
				}
				// Posting to Mastodon
				if !(config.TestMode) {
					mastoPost := mastodon.Toot{
						Status:   mastoUtil.FormatPost(track),
						MediaIDs: mediaIDs,
					}
					toot, err := mastoClient.PostStatus(ctx, &mastoPost)
					if err != nil {
						log.Printf("Error posting to Mastodon: %#v\n", err)
					}
					log.Println("Posted: ", toot.Content)
				}
				// Saving persistence data
				lastuts.LastUTS = track.Date.UTS
				state.Save(persistFile, lastuts)
			}
		}
	}
}
