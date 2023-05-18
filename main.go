/*
A music bot for discord that uses slash commands

author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/Ar5h71/r4-music-bot/bot"
	"github.com/Ar5h71/r4-music-bot/musicmanager"
)

var (
	botToken      string
	youtubeAPIKey string
)

func init() {
	flag.StringVar(&botToken, "bottoken", "", "Token for discord bot")
	flag.StringVar(&youtubeAPIKey, "youtubeapikey", "", "API key for youtube APIs")
}

func main() {
	log.Printf("************************Starting Bot************************")

	flag.Parse()
	if botToken == "" {
		log.Panicf("Please enter bot token")
	}
	if youtubeAPIKey == "" {
		log.Panicf("Please enter youtube api key")
	}
	// init youtube service client
	err := musicmanager.InitYoutubeClient(youtubeAPIKey)
	if err != nil {
		log.Panicf("Failed to init youtube client. Got error: [%s]", err.Error())
	}

	// start session for bot
	err = bot.StartBot(botToken)
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
