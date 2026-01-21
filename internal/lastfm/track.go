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
	Image []struct {
		URL  string `json:"#text"`
		Size string `json:"size"`
	}
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

func (track *Track) BestImageURL() string {
	var imageSize = map[string]int{
		"small":	1,
		"medium":	2,
		"large":	3,
		"extralarge":	4,
		"mega":		5,
	}
	bestScore := -1
	bestURL := ""
	for _, image := range track.Image {
		if image.URL == ""{
			continue
		}
		currentScore := imageSize[image.Size]
		if currentScore > bestScore {
			bestScore = currentScore
			bestURL = image.URL
		}
	}
	return bestURL
}
