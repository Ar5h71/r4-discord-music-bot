/*
command definitions and handler functions

author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package bot

import (
	"fmt"
	"time"

	"github.com/Ar5h71/r4-music-bot/common"
	"github.com/bwmarrin/discordgo"
)

// command name constants
const (
	PlayCommand      = "play"
	PlayNowCommand   = "play-now"
	PauseCommand     = "pause"
	SkipCommand      = "skip"
	ShowQueueCommand = "show-queue"
	StopQueueCommand = "stop-queue"
	ResumeCommand    = "resume"
	SearchCommand    = "search"
	AutofillCommand  = "autofill"
)

// option name constants
const (
	SongQueryOrUrlOptionName = "song-query-or-url"
	SongQueryOptionName      = "song-query"
	TimestampOptionName      = "timestamp"
	SongNumOption            = "song-num"
)

// constants for responses
const (
	InternalServerError = "Internal Server Error"
	SkipTrack           = "Skipping current playing track"
	PauseTrack          = "Pausing current playing track"
	ResumeTrack         = "Resuming current paused track"
	StopQueue           = "Stopping queue. Removing all tracks"
	ShowQueue           = "Checking all songs in queue"
	Autofill            = "Successfully generated playlist"
)

// constants for search command
const (
	SearchComponent    = "search_component"
	searchSelectHeader = "Please select a track to be added to queue"
)

// general constants
const (
	DefaultSongsForAutofill = 20
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
					Name:        SongQueryOrUrlOptionName,
					Description: "Query or URL for song to be played.",
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
					Name:        SongQueryOrUrlOptionName,
					Description: "Query or URL for song to be played.",
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
			Name:        StopQueueCommand,
			Description: "Empty the queue and stop current playing song.",
		},
		{
			Name:        ResumeCommand,
			Description: "Resume current paused song",
		},
		{
			Name:        SearchCommand,
			Description: "Search for a song",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        SongQueryOptionName,
					Description: "Query for song to be searched",
					Required:    true,
				},
			},
		},
		{
			Name:        AutofillCommand,
			Description: "Add relevant songs to queue. Songs searched based on current playing song. Current queue is ended",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        SongQueryOrUrlOptionName,
					Description: "Query or URL for song to be played and queue generated",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        SongNumOption,
					Description: "Number of songs to be added to queue",
					Required:    false,
				},
			},
		},
	}

	// command handlers for command definitions
	commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
		PlayCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			song, err := PlayCommandHandler(session, interaction, false)

			if err != nil {
				msg := fmt.Sprintf("`%s`", err.Error())
				session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
					Content: &msg,
				})
				return
			}

			addToQueueInteractionResponse(session, interaction, song, false)
		},
		PlayNowCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			song, err := PlayCommandHandler(session, interaction, true)

			if err != nil {
				msg := common.Boldify(err.Error())
				session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
					Content: &msg,
				})
				return
			}

			addToQueueInteractionResponse(session, interaction, song, true)
		},
		PauseCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			err := PauseCommandHandler(session, interaction)

			if err != nil {
				msg := common.Boldify(err.Error())
				session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
					Content: &msg,
				})
				return
			}

			msg := common.Boldify(PauseTrack)
			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
		},
		SkipCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			err := SkipCommandHandler(session, interaction)
			if err != nil {
				msg := common.Boldify(err.Error())
				session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
					Content: &msg,
				})
				return
			}

			msg := common.Boldify(SkipTrack)
			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
		},
		ShowQueueCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})

			botInstance, songs, err := ShowQueueHandler(session, interaction)
			if err != nil {
				msg := common.Boldify(err.Error())
				session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
					Content: &msg,
				})
				return
			}

			msg := ShowQueue

			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
			currentQueueMsgPaginated := generateCurrentQueueMessagePaginated(songs)

			for _, queueMsgPage := range currentQueueMsgPaginated {
				sendMessageToChannel(botInstance, queueMsgPage)
			}
		},
		StopQueueCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})

			err := EmptyQueueHandler(session, interaction)
			if err != nil {
				msg := common.Boldify(err.Error())
				session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
					Content: &msg,
				})
				return
			}

			msg := common.Boldify(StopQueue)

			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
		},
		ResumeCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			err := ResumeCommandHandler(session, interaction)

			if err != nil {
				msg := common.Boldify(err.Error())
				session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
					Content: &msg,
				})
				return
			}

			msg := common.Boldify(ResumeTrack)
			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
		},
		SearchCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags: 1000000,
				},
			})
			songs, err := SearchCommandHandler(session, interaction)
			if err != nil {
				msg := err.Error()
				session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
					Content: &msg,
				})
				return
			}

			// store searched songs in searchResults map to avoid duplicate api call
			searchResults[fmt.Sprintf("%s_%s", interaction.GuildID, interaction.Member.User.ID)] = songs

			// send interaction response for search results
			sendSearchResultsContentAndSelect(session, interaction, songs)
		},
		AutofillCommand: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			botInstance, songs, err := AutofillCommandHandler(session, interaction)

			if err != nil {
				msg := common.Boldify(err.Error())
				session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
					Content: &msg,
				})
				return
			}

			msg := Autofill

			session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
				Content: &msg,
			})
			currentQueueMsgPaginated := generateCurrentQueueMessagePaginated(songs)

			// wait for song to be played
			for botInstance.Queue.nowPlaying == nil {
				continue
			}

			time.Sleep(1 * time.Second)

			for _, queueMsgPage := range currentQueueMsgPaginated {
				sendMessageToChannel(botInstance, queueMsgPage)
			}

		},
	}
	componentHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
		SearchComponent: func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
			song, err := SearchComponentHandler(session, interaction)

			if err != nil {
				msg := common.Boldify(InternalServerError)
				session.InteractionResponseEdit(interaction.Interaction,
					&discordgo.WebhookEdit{
						Content: &msg,
					})
				return
			}

			addToQueueInteractionResponse(session, interaction, song, false)
		},
	}
	RegisteredCommands = make([]*discordgo.ApplicationCommand, len(commands))
)
