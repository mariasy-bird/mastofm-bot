package mastodonutil

import (
	"context"
	"fmt"
	"io"
	"mastofm-bot/internal/lastfm"
	"net/http"
)

func FormatPost(track *lastfm.Track) string {
	if track.Album.Text != "" {
		post := "🎵 Now listening\n" +
			track.Artist.Text + " - " + track.Name +
			"\n 📀 " + track.Album.Text
		return post
	}
	// Sometimes album is missing.
	post := "🎵 Now listening\n" +
		track.Artist.Text + " - " + track.Name +
		"\n 📀 Unknown Album"
	return post
}

// Functionality to download image, return as an array of bytes
func DownloadImage(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("URL returned %d when attempting to fetch image", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
