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

	"github.com/Ar5h71/r4-music-bot/utils"
)

// Config serializable to json: Path: "./config.json"
type ConfigStruct struct {
	BotToken string `json:"BotToken"`
}

var Config *ConfigStruct

func InitConfig() error {
	// read config file
	log.Println("Initializing config")
	file, err := ioutil.ReadFile(utils.ConfigPath)
	if err != nil {
		log.Printf("Failed to reaf config file [%s]. Got error: [%s]",
			utils.ConfigPath, err.Error())
		return err
	}

	// unmarshal file cotents to json
	Config = &ConfigStruct{}
	err = json.Unmarshal(file, Config)
	if err != nil {
		log.Printf("Failed top unmarshal json. Got error [%s]", err.Error())
		return err
	}

	log.Printf("Successfully initialized config with contents: [%#v]", RedactSensitiveDataFromConfig(*Config))
	return nil
}

func RedactSensitiveDataFromConfig(conf ConfigStruct) ConfigStruct {
	// redact bot token, keep on;y 1st 4 and last 4 chars intact
	RedactedBotToken := []rune(conf.BotToken)
	for i, _ := range RedactedBotToken {
		if i > 3 && i < len(RedactedBotToken)-4 {
			RedactedBotToken[i] = '*'
		}
	}
	conf.BotToken = string(RedactedBotToken)
	return conf
}
