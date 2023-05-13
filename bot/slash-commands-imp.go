package bot

import (
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Ar5h71/r4-music-bot/common"
	"github.com/Ar5h71/r4-music-bot/musicmanager"
	"github.com/bwmarrin/discordgo"
)

func PlayCommandHandler(session *discordgo.Session, interaction *discordgo.InteractionCreate) (*common.Song, error) {
	options := interaction.ApplicationCommandData().Options
	guildId := interaction.GuildID
	vChannelId := SearchVoiceChannelId(interaction.Member.User.ID)
	log.Printf("Guild ID: %s, vChannel ID: %s", guildId, vChannelId)
	if vChannelId == "" {
		return nil, errors.New("user needs to be in voice channel")
	}
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, option := range options {
		optionMap[option.Name] = option
	}
	log.Printf("'Play' command received")

	if _, ok := optionMap[SongQueryOptionName]; !ok {
		return nil, errors.New("You need to specify song query to play a song")
	}
	option := optionMap[SongQueryOptionName]

	log.Printf("Got option: [%s]", option.StringValue())

	// create bot instance and connect to voice channel if not there

	// make bot instance for guild if not present
	if botInstance, ok := BotInstances[guildId]; !ok {
		botVoiceConnection, err := session.ChannelVoiceJoin(guildId, vChannelId, false, false)
		if err != nil {
			log.Printf("Failed to join voice channel with id [%s]. Got error: [%s]", vChannelId, err.Error())
			return nil, errors.New("Failed to join voice channel. Internal server error")
		}
		newBotInstance := NewBotInstance(session, guildId, interaction.ChannelID, vChannelId, false, botVoiceConnection)
		BotInstances[guildId] = newBotInstance
		defer botVoiceConnection.Close()
	} else if botInstance.BotVoiceConnection == nil {
		botVoiceConnection, err := session.ChannelVoiceJoin(guildId, vChannelId, false, false)
		if err != nil {
			log.Printf("Failed to join voice channel with id [%s]. Got error: [%s]", vChannelId, err.Error())
			return nil, errors.New("Failed to join voice channel. Internal server error")
		}
		botInstance.BotVoiceConnection = botVoiceConnection
		defer botVoiceConnection.Close()
	} else {
		// if instance already exists check if bot present in same voice channel
		if vChannelId != botInstance.VoiceChannelId {
			channel, err := session.Channel(botInstance.VoiceChannelId)
			if err != nil {
				errMsg := "Failed to get current voice channel info."
				log.Printf("%s, got error: [%s]", errMsg, err.Error())
				return nil, errors.New(errMsg)
			}
			return nil, fmt.Errorf("You must be in '%s' voice channel", channel.Name)
		}
		// check if command issued from same text channel
		if interaction.ChannelID != botInstance.TextChannelId {
			channel, err := session.Channel(botInstance.TextChannelId)
			if err != nil {
				errMsg := "Failed to get text channel info"
				log.Printf("%s, got error: [%s]", errMsg, err.Error())
				return nil, errors.New(errMsg)
			}
			return nil, fmt.Errorf("You must be in '%s' text channel to issue this command", channel.Name)
		}
	}

	botInstance := BotInstances[guildId]

	songs, err := musicmanager.YtServiceClient.Search(option.StringValue(), interaction.Member.User.Username, 1)

	if err != nil {
		errMsg := "Couldn't find the requested song"
		log.Printf("%s, error: [%s]", errMsg, err.Error())
		return nil, fmt.Errorf(errMsg)
	}
	botInstance.BotVoiceConnection.LogLevel = discordgo.LogWarning
	// add to queue here
	// for now play it to check if audio is received
	botInstance.BotVoiceConnection.Speaking(true)
	defer botInstance.BotVoiceConnection.Speaking(false)
	done := make(chan error)
	botInstance.AudioStream = NewAudioStream(songs[0], botInstance.BotVoiceConnection, done)

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case err := <-done:
			if err != nil && err != io.EOF {
				// send stop signal to Audio stream
				log.Printf("Error while sending packets: %s", err.Error())
				botInstance.AudioStream.stop <- true
				return songs[0], err
			}
			log.Printf("finished playing")
			return songs[0], nil
		case <-ticker.C:
			log.Printf("Playing: packets sent: %d", botInstance.AudioStream.framesSent)
		}
	}
}
