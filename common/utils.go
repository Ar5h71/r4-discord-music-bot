/*
Package for general purpose constants and functions

author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package common

import (
	"encoding/json"
	"log"
)

const (
	ConfigPath = "config/config.json"
	BotPrefix  = "Bot "
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
