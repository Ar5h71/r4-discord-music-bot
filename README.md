# A discord music bot

A multi server discord music bot made using golang

**Features**:

- Ability to add songs using queries or youtube URLs.
- A queue to manage multiple songs.
- Pause, resume and skip functionalities for the queue.

## Steps to use

- Add your `bot token` and `youtube api key` and in `config.json`
- Run command `docker build -t <image-name>` to build the project. Replace `<image-name>` with any image name you want to give.
- Run the bot using the command `docker run <image-name>`. To run in detached mode, use `docker run -d <image-name>`

## References

- https://github.com/bwmarrin/discordgo
- https://github.com/bwmarrin/dgvoice
- https://github.com/layeh/gopus
- https://github.com/ljgago/MusicBot
