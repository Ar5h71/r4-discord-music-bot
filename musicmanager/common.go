// common vars and structs for musicmanager package
package musicmanager

// audio specifications for streaming
const (
	channels   int = 2     // 1 for mono, 2 for stereo
	frameRate  int = 48000 // audio sampling rate
	frameSize  int = 960   // uint16 size of each audio frame 960/48KHz = 20ms
	bufferSize int = 1024  // max size of opus data 1K
)
