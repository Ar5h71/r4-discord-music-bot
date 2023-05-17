package bot

import (
	"errors"
	"fmt"
	"log"
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

	// search youtube for song
	songs, err := musicmanager.YtServiceClient.Search(option.StringValue(), interaction.Member.User.Username, 1)

	if err != nil {
		errMsg := fmt.Sprintf("Couldn't find the song for query '%s'", option.StringValue())
		log.Printf("%s, error: [%s]", errMsg, err.Error())
		return nil, fmt.Errorf(errMsg)
	}
	botInstance.BotVoiceConnection.LogLevel = discordgo.LogWarning
	// send signal to songsig channel
	songSig <- &SongSignal{
		song:        songs[0],
		botInstance: botInstance,
		playNow:     playNow,
	}
	// send skip signal if playNow is true
	if playNow {
		botInstance.Queue.skip <- nil
	}
	return songs[0], nil
}

func PlayUrlCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate, playNow bool) (*common.Song, error) {
	options := interaction.ApplicationCommandData().Options
	guildId := interaction.GuildID
	vChannelId := SearchVoiceChannelId(interaction.Member.User.ID)
	logCtx := fmt.Sprintf("[%s | %s]", guildId, vChannelId)
	log.Printf("%s 'Play-url' command received", logCtx)

	// create bot instance and connect to voice channel if not there
	botInstance, err := createAndGetBotInstance(session, interaction, true)
	if err != nil {
		return nil, err
	}

	// extract option values
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, option := range options {
		optionMap[option.Name] = option
	}

	if _, ok := optionMap[SongUrlOption]; !ok {
		return nil, errors.New("You need to specify a URL to play a song")
	}
	option := optionMap[SongUrlOption]

	log.Printf("%s Got option: [%s]", logCtx, option.StringValue())

	// search youtube for song
	song, err := musicmanager.GetSongWithStreamUrl(option.StringValue(), interaction.Member.User.Username)
	if err != nil {
		log.Printf("%s Failed to get song stream url for youtube url '%s'. Got error: %s", logCtx, option.StringValue(), err.Error())
		return nil, fmt.Errorf("Couldn't find song for url '%s'", option.StringValue())
	}

	if err != nil {
		errMsg := "Couldn't find the requested song"
		log.Printf("%s, error: [%s]", errMsg, err.Error())
		return nil, fmt.Errorf(errMsg)
	}
	botInstance.BotVoiceConnection.LogLevel = discordgo.LogWarning
	// send signal to songsig channel
	songSig <- &SongSignal{
		song:        song,
		botInstance: botInstance,
		playNow:     playNow,
	}
	// send skip signal if playNow is true
	if playNow {
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
	log.Printf("[%s | %s]. 'skip' command received", guildId, vChannelId)

	botInstance, err := createAndGetBotInstance(session, interaction, false)
	if err != nil {
		return err
	}

	// send stop signal to queue
	botInstance.Queue.stop <- nil
	return nil
}

func ShowQueueHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) ([]*common.Song, error) {
	guildId := interaction.GuildID
	vChannelId := SearchVoiceChannelId(interaction.Member.User.ID)
	log.Printf("[%s | %s]. 'skip' command received", guildId, vChannelId)

	botInstance, err := createAndGetBotInstance(session, interaction, false)
	if err != nil {
		return nil, err
	}

	songs := make([]*common.Song, 0)
	songs = append(songs, botInstance.Queue.nowPlaying.song)
	songs = append(songs, botInstance.Queue.songs...)
	return songs, nil
}
