/*
command definitions and handler functions

author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package bot

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

// command name constants
const (
	PlayCommand       = "play"
	PauseCommand      = "pause"
	PlayNowCommand    = "play-now"
	SeekCommand       = "seek"
	SkipCommand       = "skip"
	ShowQueueCommand  = "show-queue"
	EmptyQueueCommand = "empty-queue"
	LeaveCommand      = "leave"
)

// option name constants
const (
	SongQueryOptionName = "song-query"
	TimestampOptionName = "timestamp"
)

var (
	// commands need to defined in slice of 'ApplicationCommand' struct
	// check 'https://github.com/bwmarrin/discordgo/blob/master/examples/slash_commands/main.go'
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        PlayCommand,
			Description: "Play a song. Add it to queue if a song is playing.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        SongQueryOptionName,
					Description: "Query for song to be played.",
					Required:    true,
				},
			},
		},
		{
			Name:        PlayNowCommand,
			Description: "Skip current playing song and play the queried song instead.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        SongQueryOptionName,
					Description: "Query for song to be played.",
					Required:    true,
				},
			},
		},
		{
			Name:        SeekCommand,
			Description: "Jump to specific timestamp in current playing song. Input format: mm:ss",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        TimestampOptionName,
					Description: "Timestamp to which to jump to.",
					Required:    true,
				},
			},
		},
		{
			Name:        PauseCommand,
			Description: "Pause currently playing song. Do nothing if no song playing.",
		},
		{
			Name:        SkipCommand,
			Description: "Skip current playing song.",
		},
		{
			Name:        ShowQueueCommand,
			Description: "Show all songs in queue.",
		},
		{
			Name:        EmptyQueueCommand,
			Description: "Empty the queue and stop current playing song.",
		},
		{
			Name:        LeaveCommand,
			Description: "Empty queue and leave the voice channel",
		},
	}

	// command handlers for command definitions
	commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
		PlayCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			// DO SOMETHING
			// dummy response for testing
			// handling for multiple options
			options := interaction.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, option := range options {
				optionMap[option.Name] = option
			}
			log.Printf("'Play' command received")

			// dummy message response for command
			msgFmt := "'Play' command received with option value: "
			if option, ok := optionMap[SongQueryOptionName]; ok {
				msgFmt += option.StringValue()
				log.Printf("Command: 'Play', Option: song-query, Value: '%s'", option.StringValue())
			}
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msgFmt,
				},
			})
		},
		PlayNowCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			// DO SOMETHING
			// dummy response for testing
			options := interaction.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, option := range options {
				optionMap[option.Name] = option
			}
			log.Printf("'Play-now' command received")

			// dummy message response for command
			msgFmt := "'Play-now' command received with option value: "
			if option, ok := optionMap[SongQueryOptionName]; ok {
				msgFmt += option.StringValue()
				log.Printf("Command: 'Play-now', Option: song-query, Value: '%s'", option.StringValue())
			}
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msgFmt,
				},
			})
		},
		SeekCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			// DO SOMETHING
			// dummy response for testing
			options := interaction.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, option := range options {
				optionMap[option.Name] = option
			}

			log.Printf("'Seek' command received")

			// dummy message response for command
			msgFmt := "'Seek' command received with option value: "
			if option, ok := optionMap[TimestampOptionName]; ok {
				msgFmt += option.StringValue()
				log.Printf("Command: 'Seek', Option: timestamp, Value: '%s'", option.StringValue())
			}
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msgFmt,
				},
			})
		},
		PauseCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			// DO SOMETHING
			// dummy response for testing

			// dummy message response for command
			msgFmt := "'Pause' command received with option value: "
			log.Printf("'Pause' command received")

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msgFmt,
				},
			})
		},
		SkipCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			// DO SOMETHING
			// dummy response for testing

			// dummy message response for command
			msgFmt := "'Skip' command received with option value: "
			log.Printf("'Skip' command received")

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msgFmt,
				},
			})
		},
		ShowQueueCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			// DO SOMETHING
			// dummy response for testing

			// dummy message response for command
			msgFmt := "'Show-queue' command received with option value: "
			log.Printf("'Show-queue' command received")

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msgFmt,
				},
			})
		},
		EmptyQueueCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			// DO SOMETHING
			// dummy response for testing

			// dummy message response for command
			msgFmt := "'Empty-queue' command received with option value: "
			log.Printf("'Empty-queue' command received")

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msgFmt,
				},
			})
		},
		LeaveCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			// DO SOMETHING
			// dummy response for testing

			// dummy message response for command
			msgFmt := "'Leave' command received with option value: "
			log.Printf("'Leave' command received")

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msgFmt,
				},
			})
		},
	}
)
