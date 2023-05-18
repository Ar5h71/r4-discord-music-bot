/*
author: Arshdeep Singh
E-mail: ad.sigh.arsh@gmail.com
*/

package bot

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"sync"

	"github.com/Ar5h71/r4-music-bot/common"
	"github.com/bwmarrin/discordgo"
	"layeh.com/gopus"
)

type AudioStreamSession struct {
	// mutex to prevent overwriting data in struct
	mtx  sync.Mutex
	song *common.Song

	done       chan error
	stop       chan interface{}
	voice      *discordgo.VoiceConnection
	framesSent int
	paused     bool
	running    bool
	err        error
}

var (
	framerate     = 48000
	framesize     = 960
	frameduration = 20
	numChannels   = 2
	maxBytes      = framesize * (frameduration / 20) * numChannels
)

func NewAudioStream(song *common.Song, voice *discordgo.VoiceConnection, done chan error) *AudioStreamSession {
	log.Printf("[%s(%s)]: Creating new stream session for song with url '%s'", song.SongTitle, song.SongId, song.SongUrl)
	audioStream := &AudioStreamSession{
		song:   song,
		voice:  voice,
		done:   done,
		paused: false,
		stop:   make(chan interface{}),
	}

	go audioStream.stream()
	return audioStream
}

func (audioStream *AudioStreamSession) stream() {
	if audioStream.running {
		log.Printf("[%s(%s)]: Stream already running for song", audioStream.song.SongTitle, audioStream.song.SongId)
		return
	}

	args := []string{
		"-i", audioStream.song.SongUrl,
		"-f", "s16le",
		"-ar", strconv.Itoa(int(framerate)),
		"-ac", strconv.Itoa(int(numChannels)),
		"-frame_duration", strconv.Itoa(int(frameduration)),
		"pipe:1",
	}

	run := exec.Command("ffmpeg", args...)

	ffmpegOut, err := run.StdoutPipe()
	if err != nil {
		audioStream.err = err
		log.Printf("[%s(%s)]: Failed to create stdout pipe for buffer. Got error: %s", audioStream.song.SongTitle, audioStream.song.SongId, err.Error())
		audioStream.done <- err
		return
	}
	ffmpegbuf := bufio.NewReaderSize(ffmpegOut, 16348)

	// start the command
	audioStream.err = run.Start()
	if err != nil {
		log.Printf("[%s(%s)]: Failed to start ffmpeg command. Error: [%s]", audioStream.song.SongTitle, audioStream.song.SongId, err.Error())
		audioStream.done <- err
		return
	}
	defer run.Process.Kill()
	go func() {
		// kill the process if stop received
		<-audioStream.stop
		audioStream.done <- run.Process.Kill()
	}()

	// channels to send packets to discord
	sendbuf := make(chan []int16, 2)
	close := make(chan interface{})

	go func() {
		SendPCMPacket(fmt.Sprintf("[%s(%s)]", audioStream.song.SongTitle, audioStream.song.SongId), audioStream.voice, sendbuf)
		close <- true
	}()

	// start reading data from stdout
	for {
		// check if stream is paused
		if audioStream.paused {
			continue
		}
		audioBuf := make([]int16, framesize*numChannels)
		err = binary.Read(ffmpegbuf, binary.LittleEndian, &audioBuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			audioStream.done <- err
			audioStream.err = err
			return
		}
		if err != nil {
			log.Printf("[%s(%s)]Failed to read audio buffer: error: [%s]", audioStream.song.SongTitle, audioStream.song.SongId, err.Error())
			audioStream.err = err
			audioStream.done <- err
			return
		}
		// Send received PCM to the sendPCM channel
		select {
		case sendbuf <- audioBuf:
			audioStream.mtx.Lock()
			audioStream.framesSent += 2
			audioStream.mtx.Unlock()
		case <-close:
			audioStream.done <- nil
			return
		}

	}

}

func SendPCMPacket(logCtx string, voice *discordgo.VoiceConnection, buf <-chan []int16) {
	if buf == nil {
		return
	}

	var err error

	opusEncoder, err := gopus.NewEncoder(int(framerate), int(numChannels), gopus.Audio)

	if err != nil {
		log.Printf("%s NewEncoder Error: %s", logCtx, err.Error())
		return
	}

	for {

		// read pcm from chan, exit if channel is closed.
		recv, ok := <-buf
		if !ok {
			log.Printf("%s PCM Channel closed", logCtx)
			return
		}

		// try encoding pcm frame with Opus
		opus, err := opusEncoder.Encode(recv, int(framesize), int(maxBytes))
		if err != nil {
			log.Printf("%s Encoding Error %s", logCtx, err.Error())
			return
		}

		if voice.Ready == false || voice.OpusSend == nil {
			// OnError(fmt.Sprintf("Discordgo not ready for opus packets. %+v : %+v", v.Ready, v.OpusSend), nil)
			// Sending errors here might not be suited
			return
		}
		// send encoded opus data to the sendOpus channel
		voice.OpusSend <- opus
	}
}

// pause ongoing stream
func (audioStream *AudioStreamSession) pauseStream() {
	audioStream.mtx.Lock()
	defer audioStream.mtx.Unlock()
	audioStream.paused = true
}

// resume ongoing stream
func (audioStream *AudioStreamSession) resumeStream() {
	audioStream.mtx.Lock()
	defer audioStream.mtx.Unlock()
	audioStream.paused = false
}
