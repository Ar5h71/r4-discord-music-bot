/*
author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package bot

import (
	"io"
	"log"
	"time"

	"github.com/Ar5h71/r4-music-bot/common"
	"github.com/Ar5h71/r4-music-bot/musicmanager"
)

// function to play the queue
func QueueInit(recv <-chan *SongSignal) {
	log.Printf("Started thread for queues")
	for {
		select {
		case songSigRecv := <-recv:
			log.Printf("[%s(%s)] Adding to queue for bot in guild (%s) and vchannel (%s)",
				songSigRecv.song.SongTitle, songSigRecv.song.SongId, songSigRecv.botInstance.GuildId,
				songSigRecv.botInstance.VoiceChannelId)
			go songSigRecv.botInstance.playQueue(songSigRecv.song, songSigRecv.playNow)
		}
	}
}

func (botInstance *BotInstance) playQueue(song *common.Song, playnow bool) {
	if playnow {
		// add the song to queue front
		botInstance.addSongFront(song)
	} else {
		botInstance.addSongBack(song)
	}
	if botInstance.Queue.nowPlaying != nil && !botInstance.Queue.nowPlaying.finished {
		log.Printf("[%s | %s]Bot is already playing",
			botInstance.GuildId, botInstance.VoiceChannelId)
		return
	}
	go func() {
		// spawn a go routine to check if queue is empty
		ticker := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-botInstance.Queue.skip:
				// skip
				botInstance.skipSong()
			case <-botInstance.Queue.stop:
				// stop queue
				botInstance.stopQueue()
			case <-botInstance.Queue.pause:
				// pause current song
				botInstance.pauseSong()
			case <-botInstance.Queue.resume:
				// resume current song
				botInstance.resumeSong()
			case <-botInstance.Queue.done:
				// queue is finished return
				StopBotInstance(botInstance)
				return
			case <-ticker.C:
				// check if anything is playing
				// if not start playing
				// log.Printf("Inside goroutine")
				if botInstance.Queue.nowPlaying != nil && !botInstance.Queue.nowPlaying.finished {
					continue
				}

				// log.Printf("1")
				// if here means nothing is playing, start playing

				botInstance.playNext()
			}
		}
	}()

}

// Add song to queue back
func (botInstance *BotInstance) addSongBack(song *common.Song) {
	log.Printf("[%s(%s)] Adding to queue back", song.SongTitle, song.SongId)
	botInstance.Queue.mtx.Lock()
	botInstance.Queue.songs = append(botInstance.Queue.songs, song)
	botInstance.Queue.mtx.Unlock()
}

// Add song to queue front
func (botInstance *BotInstance) addSongFront(song *common.Song) {
	log.Printf("[%s(%s)] Adding to queue front", song.SongTitle, song.SongId)
	botInstance.Queue.mtx.Lock()
	botInstance.Queue.songs = append([]*common.Song{song}, botInstance.Queue.songs...)
	botInstance.Queue.mtx.Unlock()
}

// skip current playing song
func (botInstance *BotInstance) skipSong() {
	log.Printf("[%s | %s] Skipping",
		botInstance.GuildId, botInstance.VoiceChannelId)
	botInstance.Queue.mtx.Lock()
	defer botInstance.Queue.mtx.Unlock()
	if botInstance.Queue.nowPlaying == nil || botInstance.Queue.nowPlaying.finished {
		log.Printf("[%s | %s] Nothing is playing",
			botInstance.GuildId, botInstance.TextChannelId)
		sendMessageToChannel(botInstance, "No song is playing. Nothing to skip")
		return
	}
	botInstance.Queue.nowPlaying.streamSession.stop <- nil
	// make nowPlaying nil
	botInstance.Queue.nowPlaying = nil
}

// delete all songs from queue
func (botInstance *BotInstance) stopQueue() {
	log.Printf("[ %s | %s ] Stopping queue",
		botInstance.GuildId, botInstance.TextChannelId)
	// stop current playing song
	botInstance.Queue.mtx.Lock()
	defer botInstance.Queue.mtx.Unlock()
	nothingToStop := true
	if botInstance.Queue.nowPlaying != nil && !botInstance.Queue.nowPlaying.finished {
		log.Printf("[%s | %s] Stopping current song",
			botInstance.GuildId, botInstance.TextChannelId)
		botInstance.Queue.nowPlaying.streamSession.stop <- nil
		botInstance.Queue.nowPlaying.finished = true
		nothingToStop = false
	}
	if len(botInstance.Queue.songs) != 0 {
		log.Printf("[%s | %s] Removing all songs",
			botInstance.GuildId, botInstance.TextChannelId)
		botInstance.Queue.songs = make([]*common.Song, 0)
		nothingToStop = false
	}
	if nothingToStop {
		log.Printf("[%s | %s] Nothing to stop",
			botInstance.GuildId, botInstance.VoiceChannelId)
		sendMessageToChannel(botInstance, "No songs in queue. Nothing to stop")
		return
	}
}

// pause current song
func (botInstance *BotInstance) pauseSong() {
	log.Printf("[%s | %s] Pausing current song",
		botInstance.GuildId, botInstance.VoiceChannelId)
	botInstance.Queue.mtx.Lock()
	defer botInstance.Queue.mtx.Unlock()
	if botInstance.Queue.nowPlaying == nil || botInstance.Queue.nowPlaying.finished {
		log.Printf("[%s | %s] Nothing to pause",
			botInstance.GuildId, botInstance.VoiceChannelId)
		sendMessageToChannel(botInstance, "No song is playing. Nothing to pause")
		return
	}
	if botInstance.Queue.paused {
		log.Printf("[%s | %s] Already paused",
			botInstance.GuildId, botInstance.VoiceChannelId)
		sendMessageToChannel(botInstance, "Queue is already paused")
		return
	}
	botInstance.Queue.nowPlaying.streamSession.pauseStream()
	botInstance.Queue.paused = true
	// send message to channel
}

// resume a song if paused
func (botInstance *BotInstance) resumeSong() {
	log.Printf("[%s | %s] Resuming current song",
		botInstance.GuildId, botInstance.VoiceChannelId)
	botInstance.Queue.mtx.Lock()
	defer botInstance.Queue.mtx.Unlock()
	if !botInstance.Queue.paused {
		log.Printf("[%s | %s] Already playing",
			botInstance.GuildId, botInstance.VoiceChannelId)
		sendMessageToChannel(botInstance, "Queue is already playing")
		return
	}
	if botInstance.Queue.nowPlaying == nil || botInstance.Queue.nowPlaying.finished {
		log.Printf("[%s | %s] Nothing to resume",
			botInstance.GuildId, botInstance.VoiceChannelId)
		sendMessageToChannel(botInstance, "Queue is already playing. Nothing to resume")
		return
	}
	botInstance.Queue.nowPlaying.streamSession.resumeStream()
	botInstance.Queue.paused = false
}

// play next song in queue
func (botInstance *BotInstance) playNext() {
	botInstance.Speaking = true
	botInstance.BotVoiceConnection.Speaking(true)
	defer func() {
		botInstance.Speaking = false
		botInstance.BotVoiceConnection.Speaking(false)
	}()
	botInstance.Queue.mtx.Lock()
	defer botInstance.Queue.mtx.Unlock()
	if len(botInstance.Queue.songs) == 0 && (botInstance.Queue.nowPlaying == nil || botInstance.Queue.nowPlaying.finished) && !botInstance.Queue.autoplay {
		botInstance.Queue.done <- nil
		return
	}

	// check if autoplay
	var song *common.Song
	var err error
	if botInstance.Queue.autoplay {
		song, err = musicmanager.YtServiceClient.GetNextSongForAutoplay(botInstance.Queue.nowPlaying.song)
		if err != nil {
			// if trouble finding song, then go with queue
			botInstance.Queue.autoplay = false
			msg := "Failed to find next song for autoplay. Switching off autoplay"
			log.Printf("[%s | %s] %s. Got error [%s]",
				botInstance.GuildId, botInstance.VoiceChannelId, msg, err.Error())
			sendMessageToChannel(botInstance, msg)
		}
	}

	if !botInstance.Queue.autoplay {
		song = botInstance.Queue.songs[0]
		log.Printf("[%s | %s] Playing song %s",
			botInstance.GuildId, botInstance.VoiceChannelId, song.SongTitle)
		if len(botInstance.Queue.songs) == 1 {
			botInstance.Queue.songs = make([]*common.Song, 0)
		} else {
			botInstance.Queue.songs = botInstance.Queue.songs[1:]
		}
	}
	done := make(chan error)
	botInstance.Queue.nowPlaying = &NowPlaying{
		song:          song,
		streamSession: NewAudioStream(song, botInstance.BotVoiceConnection, done),
		finished:      false,
	}
	sendCurrentPlayingSongMessage(botInstance, song)

	go func() {
		// wait for done channel here
		err := <-done
		if err == nil || err == io.EOF || err == io.ErrUnexpectedEOF {
			log.Printf("[%s | %s] Finished playing %s", botInstance.GuildId,
				botInstance.VoiceChannelId, song.SongTitle)
		} else if err != nil {
			log.Printf("[%s | %s] Failed to stream %s. Got error: %s", botInstance.GuildId,
				botInstance.VoiceChannelId, song.SongTitle, err.Error())
		}
		botInstance.Queue.mtx.Lock()
		defer botInstance.Queue.mtx.Unlock()
		botInstance.Queue.nowPlaying.finished = false
	}()
}
