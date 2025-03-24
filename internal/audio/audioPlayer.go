package audio

import (
	"bytes"
	"os"
	"sync/atomic"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

type Player interface {
	BufferedSize() int
	Err() error
	IsPlaying() bool
	Pause()
	Play()
	Seek(offset int64, whence int) (int64, error)
	SetBufferSize(bufferSize int)
	SetVolume(volume float64)
	Volume() float64
	Clear() error
	IsNil() bool
}

type StaticPlayer struct {
	*oto.Player

	hasBeenClosed atomic.Bool
}

func (p *StaticPlayer) Clear() error {
	err := p.Player.Close()
	if err != nil {
		return err
	}
	p.hasBeenClosed.Store(true)
	return nil
}

func (p *StaticPlayer) IsNil() bool {
	return p.hasBeenClosed.Load()
}

func CreatePlayer(mp3FilePath string) (*Player, error) {
	fileBytes, err := os.ReadFile(mp3FilePath)
	if err != nil {
		return nil, err
	}
	fileBytesReader := bytes.NewReader(fileBytes)
	decodedMp3, err := mp3.NewDecoder(fileBytesReader)
	if err != nil {
		return nil, err
	}
	player := GetAudioContext().NewPlayer(decodedMp3)
	playerWrapper := StaticPlayer{Player: player}
	playerWrapper.hasBeenClosed.Store(false)
	publicPlayer := Player(&playerWrapper)
	return &publicPlayer, nil
}
