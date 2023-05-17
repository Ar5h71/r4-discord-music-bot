/*
Bot session command regitrations

author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package bot

import (
	"log"
	"sync"

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
	Queue              *BotQueue
	// add for queue and current playing song
}

// bot queue
type BotQueue struct {
	mtx sync.Mutex

	songs      []*common.Song
	paused     bool
	nowPlaying *NowPlaying
	skip       chan interface{}
	stop       chan interface{}
	pause      chan interface{}
	resume     chan interface{}
	done       chan interface{}
}

type NowPlaying struct {
	song          *common.Song
	streamSession *AudioStreamSession
}

// to send signal in a channel to play a song for an instance
// thread is initiated in StartBot()
type SongSignal struct {
	song        *common.Song
	botInstance *BotInstance
	playNow     bool
}

var (
	BotSession    *discordgo.Session
	BotInstances  = make(map[string]*BotInstance)
	songSig       = make(chan *SongSignal)
	searchResults = make(map[string][]*common.Song)
)

func NewBotInstance(session *discordgo.Session,
	guildId,
	tChannelId,
	vchannelId string,
	speaking bool) (*BotInstance, error) {

	voiceConnection, err := session.ChannelVoiceJoin(guildId, vchannelId, false, true)
	if err != nil {
		log.Printf("[%s | %s] Failed to create voice connection. Got error: %s", guildId, vchannelId, err.Error())
		return nil, err
	}
	return &BotInstance{
		BotSession:         session,
		GuildId:            guildId,
		VoiceChannelId:     vchannelId,
		TextChannelId:      tChannelId,
		Speaking:           speaking,
		BotVoiceConnection: voiceConnection,
		Queue: &BotQueue{
			paused: false,
			stop:   make(chan interface{}, 1),
			done:   make(chan interface{}, 1),
			skip:   make(chan interface{}, 1),
			pause:  make(chan interface{}, 1),
			resume: make(chan interface{}, 1),
			songs:  make([]*common.Song, 0),
		},
	}, nil
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
		switch interaction.Type {
		case discordgo.InteractionApplicationCommand:
			if handler, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
				handler(session, interaction)
			}
		case discordgo.InteractionMessageComponent:
			if handler, ok := componentHandlers[interaction.MessageComponentData().CustomID]; ok {
				handler(session, interaction)
			}
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

	// go routine for queue
	go QueueInit(songSig)

	return nil
}

// register commands
func RegisterCommands() error {
	log.Printf("Registering commands for the bot...")

	for idx, command := range commands {
		cmd, err := BotSession.ApplicationCommandCreate(BotSession.State.User.ID, "", command)
		if err != nil {
			log.Printf("Cannot create '%v' command: %v", command.Name, err)
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

// stop bot instance
func StopBotInstance(botInstance *BotInstance) {
	// disconnect from voice
	log.Printf("disconnecting bot")
	botInstance.BotVoiceConnection.Disconnect()
	// remove botInstance from the map
	delete(BotInstances, botInstance.GuildId)
}
