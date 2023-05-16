package bot

// import (
// 	"fmt"
// 	"log"
// 	"sync"

// 	"github.com/Ar5h71/r4-music-bot/common"
// 	"github.com/Ar5h71/r4-music-bot/config"
// 	"github.com/bwmarrin/discordgo"
// )

// type OpusStreamSession struct {
// 	// mutex to prevent overwriting data in struct
// 	sync.Mutex
// 	song *common.Song

// 	done         chan error
// 	stop         chan interface{}
// 	voice        *discordgo.VoiceConnection
// 	audioFrames  [][]byte
// 	totalFrames  int
// 	currentFrame int
// 	paused       bool
// 	streaming    bool
// 	encoded      bool
// 	EncodeErorr  error
// }

// var (
// 	frameSize = config.Config.AudioConfig.AudioSamplingRate
// 	frameRate = config.Config.AudioConfig.AudioSamplingSize
// 	channels  = config.Config.AudioConfig.Channels
// 	bitRate   = config.Config.AudioConfig.AudioBitrateKbps
// )

// func EncodeSong(song *common.Song, voice *discordgo.VoiceConnection) *OpusStreamSession {
// 	logCtx := fmt.Sprintf("[%s(%s)]", song.SongTitle, song.SongId)
// 	log.Printf("%s Encoding song", logCtx)
// 	opusStream := &OpusStreamSession{
// 		song:      song,
// 		voice:     voice,
// 		streaming: false,
// 		encoded:   false,
// 	}
// 	go opusStream.encode()
// }
