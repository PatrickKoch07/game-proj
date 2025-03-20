package ui

import (
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

// solved by game state TODO
var exitRequested bool = false

type MainMenu struct {
	playButton      *button
	exitButton      *button
	SwitchSceneFunc func()
}

func (mm MainMenu) ShouldSkipUpdate() bool {
	return true
}

func (mm MainMenu) Update() {
	// maybe some animations later
}

func (mm MainMenu) InitInstance() ([]*sprites.Sprite, bool) {
	var sprites [2]*sprites.Sprite
	creationSuccess := true

	playButton, sprite, err := CreateButton(64, 256, 512, 640)
	if err != nil {
		creationSuccess = false
		logger.LOG.Error().Err(err).Msg("")
	} else {
		mm.playButton = playButton
		mm.playButton.OnPress = mm.SwitchSceneFunc
		sprites[0] = sprite
	}

	exitButton, sprite, err := CreateButton(64, 256, 512, 756)
	if err != nil {
		creationSuccess = false
		logger.LOG.Error().Err(err).Msg("")
	} else {
		mm.exitButton = exitButton
		mm.exitButton.OnPress = almostExitGame
		mm.exitButton.OnRelease = exitGame
		sprites[1] = sprite
	}

	return sprites[:], creationSuccess
}

func WasCloseRequested() bool {
	return exitRequested
}

func exitGame() {
	exitRequested = true
	// pretty sure below should only get called on the main thread...
	// glfw.GetCurrentContext().SetShouldClose(true)
}

func almostExitGame() {
	logger.LOG.Debug().Msg("This will exit if you let go over the button!!")
}
