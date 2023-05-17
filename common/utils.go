/*
Package for general purpose constants and functions

author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package common

import (
	"encoding/json"
	"fmt"
	"log"
)

const (
	ConfigPath              = "config/config.json"
	BotPrefix               = "Bot "
	YoutubeVideoURLPrefix   = "https://www.youtube.com/watch?v="
	YoutubeChannelURLPrefix = "https://www.youtube.com/channel/"
)

// function to pretty print structs
func PrettyPrint(iface interface{}) string {
	prettyStruct, err := json.MarshalIndent(iface, "", "    ")
	if err != nil {
		log.Printf("Failed to marshal struct: %+v. Got error: [%s]", iface, err.Error())
		return ""
	}
	return string(prettyStruct)
}

// function to send formatted string of text in markdown
func Boldify(msg string) string {
	return fmt.Sprintf("**%s**", msg)
}
