package musicmanager

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Ar5h71/r4-music-bot/common"
	"github.com/Ar5h71/r4-music-bot/config"
	youtubedr "github.com/kkdai/youtube/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YTService struct {
	ytService *youtube.Service
}

var (
	YtServiceClient  = &YTService{}
	downloadClient   = &youtubedr.Client{}
	youtubeUrlPrefix = "https://www.youtube.com/watch?v="
)

// init youtube service client
func InitYoutubeClient() error {
	ctx := context.Background()
	var err error
	YtServiceClient.ytService, err = youtube.NewService(ctx, option.WithAPIKey(config.Config.YoutubeConfig.ApiKey))
	if err != nil {
		log.Printf("Failed to create youtube service. Got error: [%s]", err.Error())
		return err
	}
	return nil
}

// Search single or multiple results
func (ytservice *YTService) Search(query, userName string, resultNum int64) ([]*common.Song, error) {
	// search for the query
	ytServiceSearchListCall := ytservice.ytService.Search.List([]string{"id"})
	ytServiceSearchListCall.Q(query).Type("video").VideoCategoryId("10").MaxResults(resultNum)
	ytSearchResponse, err := ytServiceSearchListCall.Do()
	if err != nil {
		log.Printf("Failed to search for query [%s]. Got error [%s]", query, err.Error())
		return nil, err
	}
	if len(ytSearchResponse.Items) == 0 {
		log.Printf("No results found for the query: %s", query)
		return nil, fmt.Errorf("No songs found for this query")
	}
	var songs []*common.Song
	for _, item := range ytSearchResponse.Items {
		videoId := item.Id.VideoId

		ytUrl := youtubeUrlPrefix + videoId

		song, err := GetSongWithStreamUrl(ytUrl, userName)
		if err != nil {
			log.Printf("Failed to get song stream URL. Got error: %s", err.Error())
			return nil, err
		}
		songs = append(songs, song)
	}
	return songs, nil
}

func GetSongWithStreamUrl(url, userName string) (*common.Song, error) {
	// get video info from url
	videoInfo, err := downloadClient.GetVideo(url)
	if err != nil {
		log.Printf("Failed to get video info. Got error: [%s]", err.Error())
	}
	songId := videoInfo.ID
	songTitle := videoInfo.Title
	channelId := videoInfo.ChannelID
	channelName := videoInfo.Author
	formats := videoInfo.Formats.WithAudioChannels().AudioChannels(2)
	formats.Sort()
	// take the best format after sorting
	songUrl, err := downloadClient.GetStreamURL(videoInfo, &formats[0])
	if err != nil {
		log.Printf("Failed to fetch stream url for video with id '%s', title '%s'. Got error: %s",
			songId, songTitle, err.Error())
		return nil, fmt.Errorf("Couldn't find stream url for the song")
	}
	duration, _ := strconv.Atoi(formats[0].ApproxDurationMs)
	songDuration := time.Duration(duration) * time.Millisecond
	return &common.Song{
		SongUrl:       songUrl,
		SongId:        songId,
		SongTitle:     songTitle,
		SongDuration:  songDuration,
		User:          userName,
		ChannelId:     channelId,
		ChannelName:   channelName,
		YoutubeSource: true,
	}, nil
}
