/*
command definitions and handler functions

author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// command name constants
const (
	PlayCommand       = "play"
	PlayNowCommand    = "play-now"
	PlayUrlCommand    = "play-url"
	PlayNowUrlCommand = "play-now-url"
	PauseCommand      = "pause"
	SkipCommand       = "skip"
	ShowQueueCommand  = "show-queue"
	EmptyQueueCommand = "empty-queue"
	ResumeCommand     = "resume"
)

// option name constants
const (
	SongQueryOptionName = "song-query"
	TimestampOptionName = "timestamp"
	SongUrlOption       = "song-url"
)

// constants for responses
const (
	PleaseWait = "Please Wait..."
)

// constants for component handlers
const (
	SearchComponent = "search_component"
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
			Name:        PlayUrlCommand,
			Description: "Play a song from youtube url. Add it to queue if song is playing.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        SongUrlOption,
					Description: "Youtube url for song to be played.",
					Required:    true,
				},
			},
		},
		{
			Name:        PlayNowCommand,
			Description: "Skip current song and play the queried song instead.",
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
			Name:        PlayNowUrlCommand,
			Description: "Skip the current song and play the queried url instead.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        SongUrlOption,
					Description: "Youtube url for song to be played.",
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
			Name:        ResumeCommand,
			Description: "Resume current paused song",
		},
	}

	// command handlers for command definitions
	commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
		PlayCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: PleaseWait,
				},
			})
			song, err := PlayCommandHandler(session, interaction, false)

			var msg string
			if err != nil {
				msg = err.Error()
			} else {
				msg = fmt.Sprintf("Adding to queue: Title - '%s', Channel - '%s', Requested by - '%s', duration - '%s'",
					song.SongTitle, song.ChannelName, song.User, song.SongDuration.String())
			}
			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})

		},
		PlayUrlCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: PleaseWait,
				},
			})
			song, err := PlayUrlCommandHandler(session, interaction, false)

			msg := ""
			if err != nil {
				msg = err.Error()
			} else {
				msg = fmt.Sprintf("Adding to queue: Title - '%s', Channel - '%s', Requested by - '%s', duration - '%s'",
					song.SongTitle, song.ChannelName, song.User, song.SongDuration.String())
			}

			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
		},
		PlayNowCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: PleaseWait,
				},
			})
			song, err := PlayCommandHandler(session, interaction, true)

			var msg string
			if err != nil {
				msg = err.Error()
			} else {
				msg = fmt.Sprintf("Executed play-now command. Added '%s' to queue.", song.SongTitle)
			}
			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
		},
		PlayNowUrlCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: PleaseWait,
				},
			})
			song, err := PlayUrlCommandHandler(session, interaction, true)

			var msg string
			if err != nil {
				msg = err.Error()
			} else {
				msg = fmt.Sprintf("Executed play-now-url command. Added '%s' to queue.", song.SongTitle)
			}
			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
		},

		PauseCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: PleaseWait,
				},
			})
			song, err := PauseCommandHandler(session, interaction)
			var msgFmt string
			if err != nil {
				msgFmt = err.Error()
			} else {
				msgFmt = fmt.Sprintf("Paused song '%s'", song.SongTitle)
			}

			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msgFmt,
			})
		},
		SkipCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: PleaseWait,
				},
			})
			song, err := SkipCommandHandler(session, interaction)
			var msgFmt string
			if err != nil {
				msgFmt = err.Error()
			} else {
				msgFmt = fmt.Sprintf("Skipped song '%s'", song.SongTitle)
			}

			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msgFmt,
			})
		},
		ShowQueueCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: PleaseWait,
				},
			})
			msgFmt := "'Show-queue' command received with option value: "
			log.Printf("'Show-queue' command received")

			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msgFmt,
			})
		},
		EmptyQueueCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: PleaseWait,
				},
			})
			msgFmt := "'Empty-queue' command received with option value: "
			log.Printf("'Empty-queue' command received")

			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msgFmt,
			})
		},
		ResumeCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: PleaseWait,
				},
			})
			song, err := ResumeCommandHandler(session, interaction)
			var msgFmt string
			if err != nil {
				msgFmt = err.Error()
			} else {
				msgFmt = fmt.Sprintf("Resumed song '%s'", song.SongTitle)
			}

			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msgFmt,
			})
		},
	}
	RegisteredCommands = make([]*discordgo.ApplicationCommand, len(commands))
)
