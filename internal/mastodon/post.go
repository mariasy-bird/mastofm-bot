package mastodonutil

import (
	"mastofm-bot/internal/lastfm"
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
