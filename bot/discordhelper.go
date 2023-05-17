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
func addToQueueInteractionResponse(session *discordgo.Session, interaction *discordgo.InteractionCreate, song *common.Song) error {
	videoUrl := common.YoutubeVideoURLPrefix + song.SongId
	channelUrl := common.YoutubeChannelURLPrefix + song.ChannelId
	msg := fmt.Sprintf("**Adding to Queue** \n\n- Title -- [%s](<%s>)\n- Channel -- [%s](<%s>)\n- Requested By -- %s\n- Duration -- %s",
		song.SongTitle, videoUrl, song.ChannelName, channelUrl, song.User, song.SongDuration.String())
	_, err := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	})
	return err
}

// send 'current playing song' message
func sendCurrentPlayingSongMessage(botInstance *BotInstance, song *common.Song) {
	videoUrl := common.YoutubeVideoURLPrefix + song.SongId
	channelUrl := common.YoutubeChannelURLPrefix + song.ChannelId
	msg := fmt.Sprintf("**Playing** \n\n- Title -- [%s](<%s>)\n- Channel -- [%s](<%s>)\n- Requested By -- %s\n- Duration -- %s",
		song.SongTitle, videoUrl, song.ChannelName, channelUrl, song.User, song.SongDuration.String())
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
	msg := fmt.Sprintf("**Current Tracks in Queue**\n\n**Now Playing**\nTitle -- [%s](<%s>)\nChannel -- [%s](<%s>)\nRequested By -- %s\nDuration -- %s\n\n",
		songs[0].SongTitle, common.YoutubeVideoURLPrefix+songs[0].SongId, songs[0].ChannelName, common.YoutubeChannelURLPrefix+songs[0].ChannelId,
		songs[0].User, songs[0].SongDuration.String())

	if len(songs) >= 1 {

		for idx, song := range songs[1:] {
			msg += fmt.Sprintf("%d. Title -- [%s](<%s>)\nChannel -- [%s](<%s>)\nRequested By -- %s\nDuration -- %s\n\n",
				idx+1, song.SongTitle, common.YoutubeVideoURLPrefix+song.SongId, song.ChannelName, common.YoutubeChannelURLPrefix+song.ChannelId,
				song.User, song.SongDuration.String())
		}
	}
	_, err := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	})
	return err
}
