package audio

import (
	"sync"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/ebitengine/oto/v3"
)

var otoContext *oto.Context
var once *sync.Once

func init() {
	once = &sync.Once{}
}

func GetAudioContext() *oto.Context {
	if otoContext == nil {
		(*once).Do(createAudioContext)
	}
	return otoContext
}

func createAudioContext() {
	op := &oto.NewContextOptions{}
	op.SampleRate = 44100               // or 48000, apparently shouldn't use other values
	op.ChannelCount = 1                 // or 2 for stereo
	op.Format = oto.FormatSignedInt16LE // default used by go-mp3 which we'll just stick to
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		logger.LOG.Error().Msg("Audio context failed to be created.")
		// reset the context
		once = &sync.Once{}
	}
	<-readyChan
	otoContext = otoCtx
}

// should only be called from the main thread
func loadAudioContext() {
	// I don't plan on hot switching audio params mid play through right now, so stub for later
}
