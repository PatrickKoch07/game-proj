package audio

import (
	"os"
	"sync/atomic"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

type StreamPlayer struct {
	*oto.Player
	file *os.File

	hasBeenClosed atomic.Bool
}

func (p *StreamPlayer) close() error {
	err := p.Close()
	if err != nil {
		return err
	}
	err = p.file.Close()
	if err != nil {
		return err
	}
	return nil
}

func (p *StreamPlayer) Clear() error {
	err := p.close()
	if err != nil {
		return err
	}
	p.hasBeenClosed.Store(true)
	return nil
}

func (p *StreamPlayer) IsNil() bool {
	return p.hasBeenClosed.Load()
}

// This creates a file tied to the player which stays open while its being streamed.
// For things that might play be played multiple times, use the non-streaming version
func CreateStreamPlayer(mp3FilePath string) (*Player, error) {
	file, err := os.Open(mp3FilePath)
	decodedMp3, err := mp3.NewDecoder(file)
	if err != nil {
		return nil, err
	}
	player := GetAudioContext().NewPlayer(decodedMp3)
	streamPlayer := StreamPlayer{Player: player, file: file}
	streamPlayer.hasBeenClosed.Store(false)
	publicPlayer := Player(&streamPlayer)
	return &publicPlayer, nil
}
