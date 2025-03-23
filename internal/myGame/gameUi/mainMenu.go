package gameUi

import (
	"github.com/PatrickKoch07/game-proj/internal/gameState"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/scenes"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
	"github.com/PatrickKoch07/game-proj/internal/text"
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

func (mm MainMenu) Kill() {
	mm.playButton.UnsubInput()
	mm.exitButton.UnsubInput()
}

func (mm MainMenu) InitInstance() ([]scenes.GameObject, []*sprites.Sprite, bool) {
	var sprites []*sprites.Sprite = make([]*sprites.Sprite, 2)
	creationSuccess := true

	textPlaySprites, ok := text.TextToSprites("play", 584, 656, 1.75, 20)
	if !ok {
		creationSuccess = false
		logger.LOG.Error().Msg(
			"failed to make some of the play button text sprites. trying to display anyway",
		)
	}
	sprites = append(sprites, textPlaySprites...)
	playButton, sprite, err := CreateButton(64, 256, 512, 640)
	if err != nil {
		creationSuccess = false
		logger.LOG.Error().Err(err).Msg("")
	} else {
		mm.playButton = playButton
		mm.playButton.OnPress = switchScene
		sprites[0] = sprite
	}

	textExitSprites, ok := text.TextToSprites("exit", 584, 772, 1.75, 20)
	if !ok {
		creationSuccess = false
		logger.LOG.Error().Msg(
			"failed to make some of the exit button text sprites. trying to display anyway",
		)
	}
	sprites = append(sprites, textExitSprites...)
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

	textSprites, ok := text.TextToSprites("Welcome to the Game!", 192, 384, 3, 20)
	if !ok {
		creationSuccess = false
		logger.LOG.Error().Msg("failed to make some of the title text sprites. trying to display anyway")
	}
	sprites = append(sprites, textSprites...)

	return []scenes.GameObject{mm}, sprites, creationSuccess
}

func switchScene() {
	gameState.GetCurrentGameState().SetFlagValue(gameState.NextScene, int32(gameState.WorldScene))
	gameState.GetCurrentGameState().SetFlagValue(gameState.LoadingScene, int32(1))
}

func exitGame() {
	gameState.GetCurrentGameState().SetFlagValue(gameState.CloseRequested, 1)
}

func almostExitGame() {
	logger.LOG.Debug().Msg("This will exit if you let go over the button!!")
}
