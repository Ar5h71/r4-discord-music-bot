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
	YtServiceClient = &YTService{}
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
	ytServiceSearchListCall := ytservice.ytService.Search.List([]string{"id", "snippet"})
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
		videoTitle := item.Snippet.Title
		song := &common.Song{
			SongId:        videoId,
			SongTitle:     videoTitle,
			User:          userName,
			YoutubeSource: true,
		}

		// Get formats and their respective stream URLs
		downloadClient := &youtubedr.Client{}
		videoInfo, err := downloadClient.GetVideo("https://www.youtube.com/watch?v=" + song.SongId)
		if err != nil {
			log.Printf("Failed to get video info. Got error: [%s]", err.Error())
		}
		formats := videoInfo.Formats.WithAudioChannels().AudioChannels(2)
		formats.Sort()

		// take the best format after sorting
		song.SongUrl, err = downloadClient.GetStreamURL(videoInfo, &formats[0])
		if err != nil {
			log.Printf("Failed to fetch stream url for video with id '%s', title '%s'. Got error: %s",
				song.SongId, song.SongTitle, err.Error())
			return nil, fmt.Errorf("Couldn't find stream url for the song")
		}
		duration, _ := strconv.Atoi(formats[0].ApproxDurationMs)
		song.SongDuration = time.Millisecond * time.Duration(duration)
		songs = append(songs, song)
	}
	return songs, nil
}
