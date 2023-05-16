/*
Package to read config.json file
replace bot token in config file

author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/
package config

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/Ar5h71/r4-music-bot/common"
)

// Config serializable to json: Path: "./config.json"
type ConfigStruct struct {
	BotConfig     BotConfigStruct     `json:"BotConfigStruct"`
	YoutubeConfig YoutubeConfigStruct `json:"YoutubeConfigStruct"`
}

type BotConfigStruct struct {
	BotToken string `json:"BotToken"`
}

type YoutubeConfigStruct struct {
	ApiKey string `json:"ApiKey"`
}

var Config *ConfigStruct

func InitConfig() error {
	// read config file
	log.Println("Initializing config")
	file, err := ioutil.ReadFile(common.ConfigPath)
	if err != nil {
		log.Printf("Failed to reaf config file [%s]. Got error: [%s]",
			common.ConfigPath, err.Error())
		return err
	}

	// unmarshal file cotents to json
	Config = &ConfigStruct{}
	err = json.Unmarshal(file, Config)
	if err != nil {
		log.Printf("Failed top unmarshal json. Got error [%s]", err.Error())
		return err
	}

	log.Printf("Successfully initialized config with contents: [%#v]", common.PrettyPrint(RedactSensitiveDataFromConfig(*Config)))
	return nil
}

func RedactSensitiveDataFromConfig(conf ConfigStruct) ConfigStruct {
	// redact bot token, keep only 1st 4 and last 4 chars intact
	redactedBotToken := []rune(conf.BotConfig.BotToken)
	for i, _ := range redactedBotToken {
		if i > 3 && i < len(redactedBotToken)-4 {
			redactedBotToken[i] = '*'
		}
	}
	conf.BotConfig.BotToken = string(redactedBotToken)

	// redact youtube api key, keep only 1st 4 and last 4 chars intact
	redactedYoutubeApiKey := []rune(conf.YoutubeConfig.ApiKey)
	for i, _ := range redactedYoutubeApiKey {
		if i > 3 && i < len(redactedYoutubeApiKey)-4 {
			redactedYoutubeApiKey[i] = '*'
		}
	}
	conf.YoutubeConfig.ApiKey = string(redactedYoutubeApiKey)
	return conf
}
