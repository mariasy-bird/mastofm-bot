package lastfm

import (
	"mastofm-bot/internal/state"
)

// Track data
type Track struct {
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

// Deduplication of tracks
func IsNew(track *Track, lastuts state.LastUTS) bool {

	if track == nil {
		return false
	}

	if track.Date.UTS == "" {
		return false
	}

	return track.Date.UTS != lastuts.LastUTS
}
