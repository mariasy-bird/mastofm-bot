package lastfm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mastofm-bot/internal/config"
	"net/http"
	"net/url"
	"strconv"
)

// Fetch all recent tracks from Last.fm, return a Track struct
func GetRecentTrack(ctx context.Context, config *config.Config) (*Track, error) {
	u, _ := url.Parse("https://ws.audioscrobbler.com/2.0")

	u.RawQuery = url.Values{
		"method":  {"user.getRecentTracks"},
		"user":    {config.LfmUsername},
		"api_key": {config.LfmApiKey},
		"format":  {"json"},
		"limit":   {"1"},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("last.fm returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed struct {
		RecentTracks struct {
			Track []Track `json:"track"`
		} `json:"recenttracks"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	if config.IgnoreNowPlaying {
		var FilteredTracks []Track
		for _, track := range parsed.RecentTracks.Track {
			nowplaying, err := strconv.ParseBool(track.Attr.Nowplaying)
			// we don't want any now-playing tracks, so if it fails to parse then we assume it's *not* now-playing
			if err != nil || !nowplaying {
				FilteredTracks = append(FilteredTracks, track)
			}
		}
		parsed.RecentTracks.Track = FilteredTracks
	}

	if len(parsed.RecentTracks.Track) == 0 {
		return nil, fmt.Errorf("last.fm returned no tracks")
	}
	return &parsed.RecentTracks.Track[0], nil
}
