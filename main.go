/*
A music bot for discord that uses slash commands

author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/Ar5h71/r4-music-bot/bot"
	"github.com/Ar5h71/r4-music-bot/config"
	"github.com/Ar5h71/r4-music-bot/musicmanager"
)

func main() {
	log.Printf("************************Starting Bot************************")

	// setup config for bot
	err := config.InitConfig()
	if err != nil {
		log.Panicf("Failed to initialize config. Got error: [%s]", err.Error())
	}

	// init youtube service client
	err = musicmanager.InitYoutubeClient()
	if err != nil {
		log.Panicf("Failed to init youtube client. Got error: [%s]", err.Error())
	}

	// start session for bot
	err = bot.StartBot()
	if err != nil {
		log.Panicf("Failed to start bot: Got error: [%s]", err.Error())
	}
	defer bot.StopBot()

	// register commands
	err = bot.RegisterCommands()
	if err != nil {
		log.Panicf("Failed to register commands for bot: [%s]", err.Error())
	}
	defer bot.RemoveCommands()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Printf("Press Ctrl+c to stop the bot")
	<-stop

	log.Printf("Gracefully shutting down bot")
}
