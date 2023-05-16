package common

import (
	"time"
)

// struct for song to be streamed
type Song struct {
	SongUrl       string
	SongTitle     string
	SongDuration  time.Duration
	User          string
	SongId        string
	ChannelId     string
	ChannelName   string
	YoutubeSource bool
}
