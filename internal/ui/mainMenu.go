package ui

import (
	"github.com/PatrickKoch07/game-proj/internal/gameState"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

type MainMenu struct {
	playButton *button
	exitButton *button
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
		mm.playButton.OnPress = switchScene
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

func switchScene() {
	gameState.GetCurrentGameState().SetFlagValue(gameState.NextScene, int(gameState.WorldScene))
}

func exitGame() {
	gameState.GetCurrentGameState().SetFlagValue(gameState.CloseRequested, 1)
}

func almostExitGame() {
	logger.LOG.Debug().Msg("This will exit if you let go over the button!!")
}
