/*
author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package bot

import (
	"fmt"
	"log"

	"github.com/Ar5h71/r4-music-bot/common"
	"github.com/bwmarrin/discordgo"
)

func SearchVoiceChannelId(userId string) string {
	for _, guild := range BotSession.State.Guilds {
		for _, vChannel := range guild.VoiceStates {
			if vChannel.UserID == userId {
				return vChannel.ChannelID
			}
		}
	}
	return ""
}

// send 'adding to queue' message
func addToQueueInteractionResponse(session *discordgo.Session, interaction *discordgo.InteractionCreate, song *common.Song, playNow bool) error {
	videoUrl := common.YoutubeVideoURLPrefix + song.SongId
	channelUrl := common.YoutubeChannelURLPrefix + song.ChannelId
	var msg string
	if playNow {
		msg = fmt.Sprintf("**Adding to Queue Top** \n\n`%s` -- [%s](<%s>) | [%s](<%s>) | Requested by -- `%s`",
			song.SongDuration.String(), song.SongTitle, videoUrl, song.ChannelName, channelUrl, song.User)
	} else {
		msg = fmt.Sprintf("**Adding to Queue** \n\n`%s` -- [%s](<%s>) | [%s](<%s>) | Requested by -- `%s`",
			song.SongDuration.String(), song.SongTitle, videoUrl, song.ChannelName, channelUrl, song.User)
	}
	_, err := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	})
	return err
}

// send 'current playing song' message
func sendCurrentPlayingSongMessage(botInstance *BotInstance, song *common.Song) {
	msg := fmt.Sprintf("**Playing** \n\n`%s` -- `%s` | `%s` | Requested by -- `%s`",
		song.SongDuration.String(), song.SongTitle, song.ChannelName, song.User)

	_, err := botInstance.BotSession.ChannelMessageSend(botInstance.TextChannelId, msg)
	if err != nil {
		log.Printf("[%s | %s] Failed to send current playing song message. Got error: [%s]",
			botInstance.GuildId, botInstance.VoiceChannelId, err.Error())
	}
}

// send any message to discord channel
func sendMessageToChannel(botInstance *BotInstance, msg string) {
	_, err := botInstance.BotSession.ChannelMessageSend(botInstance.TextChannelId, common.Boldify(msg))
	if err != nil {
		log.Printf("[%s | %s] Failed to send message '%s'. Got error: %s",
			botInstance.GuildId, botInstance.VoiceChannelId, msg, err.Error())
	}
}

// send message for current songs in queue
func sendCurrentQueueInteractionResponse(session *discordgo.Session, interaction *discordgo.InteractionCreate, songs []*common.Song) error {
	videoUrl := fmt.Sprintf("%s%s", common.YoutubeVideoURLPrefix, songs[0].SongId)
	channelUrl := fmt.Sprintf("%s%s", common.YoutubeChannelURLPrefix, songs[0].ChannelId)
	msg := fmt.Sprintf("**Current Tracks in Queue**\n\n**Now Playing**\n%s -- [%s](<%s>) | [%s](<%s> | Requested by -- `%s`\n\n",
		songs[0].SongDuration.String(), songs[0].SongTitle, videoUrl, songs[0].ChannelName, channelUrl, songs[0].User)

	if len(songs) > 1 {

		for idx, song := range songs[1:] {
			videoUrl := fmt.Sprintf("%s%s", common.YoutubeVideoURLPrefix, song.SongId)
			channelUrl := fmt.Sprintf("%s%s", common.YoutubeChannelURLPrefix, song.ChannelId)
			msg += fmt.Sprintf("%d. `%s` -- [%s](<%s>) | [%s](<%s> | Requested by -- `%s`)\n",
				idx+1, song.SongDuration.String(), song.SongTitle, videoUrl, song.ChannelName, channelUrl, song.User)
		}
	}
	_, err := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	})
	return err
}

// send response for search results and song select
func sendSearchResultsContentAndSelect(session *discordgo.Session, interaction *discordgo.InteractionCreate, songs []*common.Song) error {
	// generate a select menu for searched songs
	var selectMenuOptions = make([]discordgo.SelectMenuOption, 0)
	for idx, song := range songs {
		selectMenuOptions = append(selectMenuOptions, discordgo.SelectMenuOption{
			Label:       song.SongTitle,
			Value:       fmt.Sprintf("%d", idx),
			Description: song.ChannelName,
		})
	}
	searchSelectMenuComponent := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.SelectMenu{
					CustomID:    SearchComponent,
					Placeholder: searchSelectHeader,
					Options:     selectMenuOptions,
				},
			},
		},
	}

	msg := "**Search Results\n\n**"
	if len(songs) > 1 {
		for idx, song := range songs {
			videoUrl := fmt.Sprintf("%s%s", common.YoutubeVideoURLPrefix, song.SongId)
			channelUrl := fmt.Sprintf("%s%s", common.YoutubeChannelURLPrefix, song.ChannelId)
			msg += fmt.Sprintf("%d. `%s` -- [%s](<%s>) | [%s](<%s>)\n",
				idx+1, song.SongDuration.String(), song.SongTitle, videoUrl, song.ChannelName, channelUrl)
		}
	}

	_, err := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Components: &searchSelectMenuComponent,
		Content:    &msg,
	})
	return err
}
