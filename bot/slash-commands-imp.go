package bot

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/Ar5h71/r4-music-bot/common"
	"github.com/Ar5h71/r4-music-bot/musicmanager"
	"github.com/bwmarrin/discordgo"
)

func createAndGetBotInstance(session *discordgo.Session, interaction *discordgo.InteractionCreate, create bool) (*BotInstance, error) {
	guildId := interaction.GuildID
	vChannelId := SearchVoiceChannelId(interaction.Member.User.ID)
	logCtx := fmt.Sprintf("[%s | %s]", guildId, vChannelId)

	// check if interaction issued from a user in voice channel
	if vChannelId == "" {
		log.Printf("%s %s user not in a voice channel", logCtx, interaction.Member.User.Username)
		return nil, errors.New("You need to be in a voice channel to use this command")
	}

	// check if bot instance already exists for the guild from where
	// interaction came
	botInstance, ok := BotInstances[guildId]
	// if create flag is false, return with error
	if !ok {
		if !create {
			log.Printf("%s, Create flag set to false. Returning", logCtx)
			return nil, errors.New("Play a song first to use this command")
		}
		log.Printf("%s Creating bot instance.", logCtx)
		botInstance, err := NewBotInstance(session, guildId, interaction.ChannelID, vChannelId, false)
		if err != nil {
			log.Printf("%s Failed to create bot instance. Got error: %s", logCtx, err.Error())
			return nil, errors.New("Failed to join voice channel. Internal server error")
		}
		// save to map
		BotInstances[guildId] = botInstance
		return botInstance, nil
	}
	// if instance already present, check if command received from correct voice
	// and text channel
	if botInstance.TextChannelId != interaction.ChannelID {
		log.Printf("%s command received from different text channel", logCtx)
		channel, err := session.Channel(botInstance.TextChannelId)
		if err != nil {
			log.Printf("%s Failed to get text channel information. Got error: %s", logCtx, err.Error())
			return nil, errors.New("Internal Server error")
		}
		return nil, fmt.Errorf("You need to be in '%s' text channel to use this command", channel.Name)
	}
	if botInstance.BotVoiceConnection == nil {
		voiceConnection, err := session.ChannelVoiceJoin(guildId, vChannelId, false, true)
		if err != nil {
			log.Printf("%s Failed to create voice connection. Got error: %s", logCtx, err.Error())
			return nil, errors.New("Failed to join voice channel. Internal server error")
		}
		botInstance.BotVoiceConnection = voiceConnection
		return botInstance, nil
	}

	// check if correct voice channel
	if botInstance.VoiceChannelId != vChannelId {
		log.Printf("%s command received from different voice channel", logCtx)
		channel, err := session.Channel(botInstance.VoiceChannelId)
		if err != nil {
			log.Printf("%s Failed to get text channel information. Got error: %s", logCtx, err.Error())
			return nil, errors.New("Internal Server error")
		}
		return nil, fmt.Errorf("You need to be in '%s' text channel to use this command", channel.Name)
	}
	return botInstance, nil
}

func PlayCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate, playNow bool) (*common.Song, error) {
	options := interaction.ApplicationCommandData().Options
	guildId := interaction.GuildID
	vChannelId := SearchVoiceChannelId(interaction.Member.User.ID)
	logCtx := fmt.Sprintf("[%s | %s]", guildId, vChannelId)
	log.Printf("%s 'Play' command received", logCtx)

	// create bot instance and connect to voice channel if not there
	botInstance, err := createAndGetBotInstance(session, interaction, true)
	if err != nil {
		return nil, err
	}

	option := options[0]

	log.Printf("%s Got option: [%s]", logCtx, option.StringValue())

	var song *common.Song

	// check if option received is url
	_, err = url.ParseRequestURI(option.StringValue())
	if err == nil {
		log.Printf("%s Received option is a URL: [%s]", logCtx, option.StringValue())
		song, err = musicmanager.GetSongWithStreamUrl(option.StringValue(), interaction.Member.User.Username)
		if err != nil {
			errMsg := fmt.Sprintf("Couldn't find song for the requested URL '%s'", option.StringValue())
			log.Printf("%s error [%s]", logCtx, err.Error())
			return nil, fmt.Errorf(errMsg)
		}
	} else {

		// search youtube for song
		songs, err := musicmanager.YtServiceClient.Search(option.StringValue(), interaction.Member.User.Username, 1)
		song = songs[0]

		if err != nil {
			errMsg := fmt.Sprintf("Couldn't find the song for query '%s'", option.StringValue())
			log.Printf("%s, error: [%s]", errMsg, err.Error())
			return nil, fmt.Errorf(errMsg)
		}
	}

	botInstance.BotVoiceConnection.LogLevel = discordgo.LogWarning
	// send signal to songsig channel
	songSig <- &SongSignal{
		song:        song,
		botInstance: botInstance,
		playNow:     playNow,
	}
	// send skip signal if playNow is true
	if playNow && botInstance.Queue.nowPlaying != nil {
		botInstance.Queue.skip <- nil
	}
	return song, nil
}

func PauseCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
	guildId := interaction.GuildID
	vChannelId := SearchVoiceChannelId(interaction.Member.User.ID)
	log.Printf("[%s | %s]. 'Pause' command received", guildId, vChannelId)

	botInstance, err := createAndGetBotInstance(session, interaction, false)
	if err != nil {
		return err
	}
	// send signal on pause channel
	botInstance.Queue.pause <- nil
	return nil
}

func ResumeCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
	guildId := interaction.GuildID
	vChannelId := SearchVoiceChannelId(interaction.Member.User.ID)
	log.Printf("[%s | %s]. 'Resume' command received", guildId, vChannelId)

	botInstance, err := createAndGetBotInstance(session, interaction, false)
	if err != nil {
		return err
	}

	// send signal on pause channel
	botInstance.Queue.resume <- nil
	return nil
}

func SkipCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
	guildId := interaction.GuildID
	vChannelId := SearchVoiceChannelId(interaction.Member.User.ID)
	log.Printf("[%s | %s]. 'skip' command received", guildId, vChannelId)

	botInstance, err := createAndGetBotInstance(session, interaction, false)
	if err != nil {
		return err
	}

	// send signal on pause channel
	botInstance.Queue.skip <- nil
	return nil
}

func SearchCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) ([]*common.Song, error) {
	options := interaction.ApplicationCommandData().Options
	guildId := interaction.GuildID
	vChannelId := SearchVoiceChannelId(interaction.Member.User.ID)
	logCtx := fmt.Sprintf("[%s | %s]", guildId, vChannelId)
	log.Printf("%s 'Search' command received", logCtx)

	option := options[0]

	log.Printf("%s Got option: [%s]", logCtx, option.StringValue())

	// search youtube for song
	songs, err := musicmanager.YtServiceClient.Search(option.StringValue(), interaction.Member.User.Username, 10)

	if err != nil {
		errMsg := fmt.Sprintf("Couldn't find the songs for query '%s'", option.StringValue())
		log.Printf("%s, error: [%s]", errMsg, err.Error())
		return nil, fmt.Errorf(errMsg)
	}

	if len(songs) == 0 {
		errMsg := fmt.Sprintf("Couldn't find the songs for query '%s'", option.StringValue())
		log.Printf("%s", errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	return songs, nil
}

func SearchComponentHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) (*common.Song, error) {
	data := interaction.MessageComponentData()
	guildId := interaction.GuildID
	userId := interaction.Member.User.ID
	vChannelId := SearchVoiceChannelId(userId)
	songIdx, err := strconv.Atoi(data.Values[0])
	if err != nil {
		log.Printf("[%s | %s] Failed to parse song index '%s' to integer. Got error [%s]",
			guildId, vChannelId, data.Values[0], err.Error())
		return nil, err
	}
	key := fmt.Sprintf("%s_%s", guildId, userId)
	songs, ok := searchResults[key]
	if !ok {
		log.Printf("[%s | %s] Failed to find songs for key '%s'",
			guildId, vChannelId, key)
		return nil, err
	}
	// delete the key from the map to avoid memory leak
	delete(searchResults, key)
	song := songs[songIdx]
	botInstance, err := createAndGetBotInstance(session, interaction, true)
	if err != nil {
		log.Printf("[%s | %s] Failed to create bot instance. Got error: [%s]",
			guildId, vChannelId, err.Error())
	}
	log.Printf("[%s | %s] Adding song '%s' to playlist",
		guildId, vChannelId, song.SongTitle)

	// send signal to add song to queue
	songSig <- &SongSignal{
		song:        song,
		botInstance: botInstance,
		playNow:     false,
	}
	return song, nil
}

func EmptyQueueHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) error {
	guildId := interaction.GuildID
	vChannelId := SearchVoiceChannelId(interaction.Member.User.ID)
	log.Printf("[%s | %s]. 'empty-queue' command received", guildId, vChannelId)

	botInstance, err := createAndGetBotInstance(session, interaction, false)
	if err != nil {
		return err
	}

	// send stop signal to queue
	botInstance.Queue.stop <- nil
	return nil
}

func ShowQueueHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) (*BotInstance, []*common.Song, error) {
	guildId := interaction.GuildID
	vChannelId := SearchVoiceChannelId(interaction.Member.User.ID)
	log.Printf("[%s | %s]. 'show-queue' command received", guildId, vChannelId)

	botInstance, err := createAndGetBotInstance(session, interaction, false)
	if err != nil {
		return botInstance, nil, err
	}

	songs := make([]*common.Song, 0)
	songs = append(songs, botInstance.Queue.nowPlaying.song)
	songs = append(songs, botInstance.Queue.songs...)
	return botInstance, songs, nil
}

// if autoplay received, stop current song, search for relevant songs, add to
// queue and start playing queried song
func AutofillCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) (*BotInstance, []*common.Song, error) {
	options := interaction.ApplicationCommandData().Options
	guildId := interaction.GuildID
	vChannelId := SearchVoiceChannelId(interaction.Member.User.ID)
	logCtx := fmt.Sprintf("[%s | %s]", guildId, vChannelId)
	log.Printf("%s 'autofill' command received", logCtx)

	// create bot instance and connect to voice channel if not there
	botInstance, err := createAndGetBotInstance(session, interaction, true)
	if err != nil {
		return nil, nil, err
	}
	var songQuery string
	var songNum int
	optionMap := make(map[string]string, 0)
	for _, option := range options {
		optionMap[option.Name] = option.StringValue()
	}

	if val, ok := optionMap[SongQueryOrUrlOptionName]; ok {
		songQuery = val
	}
	if val, ok := optionMap[SongNumOption]; ok {
		songNum, err = strconv.Atoi(val)
		if err != nil {
			log.Printf("[%s] Error when parsing songNum option. Error: %s", logCtx, err.Error())
			return botInstance, nil, fmt.Errorf("Please specify number of songs correctly.")
		}
	} else {
		songNum = DefaultSongsForAutofill
	}

	log.Printf("[%s] Got options for autoplay - %s: %s, %s: %d", logCtx, SongQueryOrUrlOptionName, songQuery, SongNumOption, songNum)

	var song *common.Song

	// check if option received is url
	_, err = url.ParseRequestURI(songQuery)
	if err == nil {
		log.Printf("%s Received option is a URL: [%s]", logCtx, songQuery)
		song, err = musicmanager.GetSongWithStreamUrl(songQuery, interaction.Member.User.Username)
		if err != nil {
			errMsg := fmt.Sprintf("Couldn't find song for the requested URL '%s'", songQuery)
			log.Printf("%s error [%s]", logCtx, err.Error())
			return botInstance, nil, fmt.Errorf(errMsg)
		}
	} else {

		// search youtube for song
		songs, err := musicmanager.YtServiceClient.Search(songQuery, interaction.Member.User.Username, 1)
		song = songs[0]

		if err != nil {
			errMsg := fmt.Sprintf("Couldn't find the song for query '%s'", songQuery)
			log.Printf("%s, error: [%s]", errMsg, err.Error())
			return botInstance, nil, fmt.Errorf(errMsg)
		}
	}

	botInstance.BotVoiceConnection.LogLevel = discordgo.LogWarning

	// search songs related to queried song
	songs, err := musicmanager.YtServiceClient.SearchRelavantSongs(song.SongId, song.User, int64(songNum))
	if err != nil {
		log.Printf("[%s] Failed to get relevant songs for song [%s | %s]", logCtx, song.SongTitle, song.SongId)
		return botInstance, nil, fmt.Errorf("Failed to generate queue")
	}

	// send signal to songsig channel to play queried song first
	songSig <- &SongSignal{
		song:        song,
		botInstance: botInstance,
		playNow:     true,
	}

	// change queue with searched songs
	botInstance.Queue.mtx.Lock()
	botInstance.Queue.songs = songs
	botInstance.Queue.mtx.Unlock()
	// send skip signal to stop current playing song
	if botInstance.Queue.nowPlaying != nil {
		botInstance.Queue.skip <- nil
	}

	songsInQueue := make([]*common.Song, 0)
	songsInQueue = append(songsInQueue, song)
	songsInQueue = append(songsInQueue, botInstance.Queue.songs...)

	return botInstance, songsInQueue, nil
}
