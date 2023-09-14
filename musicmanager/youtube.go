/*
author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package musicmanager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Ar5h71/r4-music-bot/common"
	youtubedr "github.com/kkdai/youtube/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type YTService struct {
	ytService *youtube.Service
}

var (
	YtServiceClient = &YTService{}
	downloadClient  = &youtubedr.Client{}
)

// init youtube service client
func InitYoutubeClient(youtubeAPIKey string) error {
	log.Printf("Initializing youtube client...")
	ctx := context.Background()
	var err error
	YtServiceClient.ytService, err = youtube.NewService(ctx, option.WithAPIKey(youtubeAPIKey))
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
	var errMsgs []string
	wg := new(sync.WaitGroup)
	wg.Add(len(ytSearchResponse.Items))
	for _, item := range ytSearchResponse.Items {
		go func(vidId string) {
			defer wg.Done()

			ytUrl := common.YoutubeVideoURLPrefix + vidId

			song, err := GetSongWithStreamUrl(ytUrl, userName)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to get song stream URL for song with id '%s'. Error [%s]", vidId, err.Error())
				errMsgs = append(errMsgs, errMsg)
				return
			}
			songs = append(songs, song)
		}(item.Id.VideoId)
	}
	log.Printf("Waiting for stream url for search results for query %s", query)
	wg.Wait()
	log.Printf("Fetched stream url for all search results for query %s", query)
	if len(errMsgs) > 0 {
		return nil, errors.New(strings.Join(errMsgs, ";"))
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
	if len(formats) == 0 {
		log.Printf("No formats for video with id '%s', title '%s'",
			songId, songTitle)
		return nil, fmt.Errorf("No valid formats found for the song")
	}
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

// Search single or multiple results
func (ytservice *YTService) SearchRelavantSongs(videoId, userName string, resultNum int64) ([]*common.Song, error) {
	// search for the query
	ytServiceSearchListCall := ytservice.ytService.Search.List([]string{"id"})
	ytServiceSearchListCall.Type("video").VideoCategoryId("10").MaxResults(resultNum).RelatedToVideoId(videoId).VideoDuration("short").VideoSyndicated("any")
	ytSearchResponse, err := ytServiceSearchListCall.Do()
	if err != nil {
		log.Printf("Failed to search relevant songs for id [%s]. Got error [%s]", videoId, err.Error())
		return nil, err
	}
	if len(ytSearchResponse.Items) == 0 {
		log.Printf("No results found related to video id: %s", videoId)
		return nil, fmt.Errorf("No songs found for this query")
	}
	var songs []*common.Song
	var errMsgs []string
	wg := new(sync.WaitGroup)
	wg.Add(len(ytSearchResponse.Items))
	for _, item := range ytSearchResponse.Items {
		go func(vidId string) {
			defer wg.Done()

			ytUrl := common.YoutubeVideoURLPrefix + vidId

			song, err := GetSongWithStreamUrl(ytUrl, userName)
			if err != nil {
				log.Printf("Failed to get song stream URL. Got error: %s", err.Error())
				errMsg := fmt.Sprintf("Failed to get song stream URL for song with id '%s'. Error [%s]", vidId, err.Error())
				errMsgs = append(errMsgs, errMsg)
				return
			}
			songs = append(songs, song)
		}(item.Id.VideoId)
	}
	wg.Wait()
	if len(errMsgs) > 0 {
		return nil, errors.New(strings.Join(errMsgs, ";"))
	}
	return songs, nil
}
