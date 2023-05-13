package bot

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
