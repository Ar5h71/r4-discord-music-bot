/*
Bot session command regitrations

author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package bot

import (
	"errors"
	"log"
	"strings"
	"sync"

	"github.com/Ar5h71/r4-music-bot/config"
	"github.com/Ar5h71/r4-music-bot/utils"
	"github.com/bwmarrin/discordgo"
)

var (
	BotSession         *discordgo.Session
	RegisteredCommands = make([]*discordgo.ApplicationCommand, len(commands))
)

// store response for concurrent registering and deregistering commands
type RegisterCommandResp struct {
	commandName string
	err         error
}

// Create and open bot session
func StartBot() error {
	log.Printf("Initializing bot session.")
	var err error

	// Create session
	BotSession, err = discordgo.New(utils.BotPrefix + config.Config.BotToken)
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

	wg := new(sync.WaitGroup)
	wg.Add(len(commands))

	var registerCommandResp []*RegisterCommandResp
	for idx, command := range commands {
		go func(cmd *discordgo.ApplicationCommand, i int) {
			defer wg.Done()
			registeredCmd, err := BotSession.ApplicationCommandCreate(BotSession.State.User.ID, "", cmd)
			response := &RegisterCommandResp{
				commandName: cmd.Name,
				err:         nil,
			}
			if err != nil {
				log.Printf("Cannot create '%v' command: %v", cmd.Name, err)
				response.err = err
			}
			log.Printf("Registered command: [%s]", cmd.Name)
			RegisteredCommands[i] = registeredCmd
			registerCommandResp = append(registerCommandResp, response)
		}(command, idx)
	}
	log.Printf("waiting for all commands to get registered")
	wg.Wait()
	log.Printf("Attempted to register all commands")
	var errMsg []string
	for _, response := range registerCommandResp {
		if response.err != nil {
			errMsg = append(errMsg, response.err.Error())
		}
	}
	if len(errMsg) > 0 {
		return errors.New(strings.Join(errMsg, ","))
	}
	return nil
}

// remove commands
func RemoveCommands() error {
	log.Printf("Removing commands")

	wg := new(sync.WaitGroup)
	wg.Add(len(RegisteredCommands))
	for _, command := range RegisteredCommands {
		go func(cmd *discordgo.ApplicationCommand) {
			defer wg.Done()
			if cmd == nil {
				log.Printf("command is nil")
			}
			err := BotSession.ApplicationCommandDelete(BotSession.State.User.ID, "", cmd.ID)
			if err != nil {
				log.Printf("Cannot delete '%v' command: %v", cmd.Name, err)
			}
		}(command)
	}
	log.Printf("Attempting to remove all commands.")
	wg.Wait()
	log.Printf("Successfully attempted to removed all commands")
	return nil
}

// stop session for the bot
func StopBot() {
	if err := BotSession.Close(); err != nil {
		log.Printf("Failed to close bot session. Got error: [%s]", err.Error())
	}
}
