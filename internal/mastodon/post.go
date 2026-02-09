package mastodonutil

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"mastofm-bot/internal/lastfm"
	"net/http"

	"github.com/mattn/go-mastodon"
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
	// This logic is to detect if it's uploading null album art; the magic number is last.fm's "no album art" art
	if strings.Contains(url, "2a96cbd8b46e442fc41c2b86b821562f") {
    	return nil, fmt.Errorf("URL returned generic album art; not fetching album art")
	}
	
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

// Create a media attachment with alt text containing track info
func UploadAlbumArt(ctx context.Context, c *mastodon.Client, imgBytes []byte, track *lastfm.Track) (*mastodon.Attachment, error) {
	var media mastodon.Media
	media.File = bytes.NewReader(imgBytes)
	media.Description = fmt.Sprintf("Album cover for %s by %s.", track.Album.Text, track.Artist.Text)

	attachment, err := c.UploadMediaFromMedia(ctx, &media)
	if err != nil {
		return nil, err
	}
	return attachment, nil
}
