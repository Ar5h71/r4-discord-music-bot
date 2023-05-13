/*
Bot session command regitrations

author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package bot

import (
	"log"

	"github.com/Ar5h71/r4-music-bot/common"
	"github.com/Ar5h71/r4-music-bot/config"
	"github.com/bwmarrin/discordgo"
)

// struct for bot instance on a guild
type BotInstance struct {
	BotSession         *discordgo.Session
	BotVoiceConnection *discordgo.VoiceConnection
	GuildId            string
	VoiceChannelId     string
	TextChannelId      string
	Speaking           bool
	AudioStream        *AudioStreamSession
	// add for queue and current playing song
}

var (
	BotSession   *discordgo.Session
	BotInstances = make(map[string]*BotInstance)
)

func NewBotInstance(session *discordgo.Session,
	guildId,
	tChannelId,
	vchannelId string,
	speaking bool,
	voiceConnection *discordgo.VoiceConnection) *BotInstance {
	return &BotInstance{
		BotSession:         session,
		GuildId:            guildId,
		VoiceChannelId:     vchannelId,
		TextChannelId:      tChannelId,
		Speaking:           speaking,
		BotVoiceConnection: voiceConnection,
	}
}

// Create and open bot session and voice connection
func StartBot() error {
	log.Printf("Initializing bot session.")
	var err error

	// Create session
	BotSession, err = discordgo.New(common.BotPrefix + config.Config.BotConfig.BotToken)
	if err != nil {
		log.Printf("Failed to start new session for bot. Got error: [%s]", err.Error())
		return err
	}

	// add handler for command handlers
	BotSession.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
			handler(session, interaction)
		}
	})

	// add websocket ready event
	BotSession.AddHandler(func(session *discordgo.Session, ready *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", session.State.User.Username, session.State.User.Discriminator)
	})

	err = BotSession.Open()
	if err != nil {
		log.Printf("Failed to create websocket session for discord. Got error: [%s]", err.Error())
		return err
	}

	return nil
}

// register commands
func RegisterCommands() error {
	log.Printf("Registering commands for the bot...")

	for idx, command := range commands {
		cmd, err := BotSession.ApplicationCommandCreate(BotSession.State.User.ID, "", command)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v", cmd.Name, err)
			return err
		}
		log.Printf("Registered command: [%s]", cmd.Name)
		RegisteredCommands[idx] = cmd
	}
	log.Printf("Registered all commands")
	return nil
}

// remove commands
func RemoveCommands() error {
	log.Printf("Removing commands")

	for _, command := range RegisteredCommands {
		if command == nil {
			log.Printf("command is nil")
			continue
		}
		err := BotSession.ApplicationCommandDelete(BotSession.State.User.ID, "", command.ID)
		if err != nil {
			log.Printf("Cannot delete '%v' command: %v", command.Name, err)
		}
	}
	log.Printf("Successfully removed all commands")
	return nil
}

// stop session for the bot
func StopBot() {
	if err := BotSession.Close(); err != nil {
		log.Printf("Failed to close bot session. Got error: [%s]", err.Error())
	}
}
