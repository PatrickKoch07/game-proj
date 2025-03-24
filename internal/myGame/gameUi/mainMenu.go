package gameUi

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/audio"
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

func (mm MainMenu) IsDead() bool {
	return false
}

func (mm MainMenu) Kill() {
	mm.playButton.UnsubInput()
	mm.exitButton.UnsubInput()
}

func (mm MainMenu) InitInstance() ([]scenes.GameObject, []*sprites.Sprite, []*audio.Player, bool) {
	var Sprites []*sprites.Sprite = make([]*sprites.Sprite, 2)
	var AudioPlayers []*audio.Player = make([]*audio.Player, 2)
	creationSuccess := true

	textPlaySprites, ok := text.TextToSprites("play", 584, 656, 1.75, 20)
	if !ok {
		creationSuccess = false
		logger.LOG.Error().Msg(
			"failed to make some of the play button text sprites. trying to display anyway",
		)
	}
	Sprites = append(Sprites, textPlaySprites...)
	playButton, err := CreateButton(64, 256, 512, 640)
	if err != nil {
		creationSuccess = false
		logger.LOG.Error().Err(err).Msg("")
	} else {
		mm.playButton = playButton
		mm.playButton.OnPress = switchScene
		Sprites[0] = mm.playButton.Sprite
		AudioPlayers[0] = mm.playButton.AudioPlayer
	}

	textExitSprites, ok := text.TextToSprites("exit", 584, 772, 1.75, 20)
	if !ok {
		creationSuccess = false
		logger.LOG.Error().Msg(
			"failed to make some of the exit button text sprites. trying to display anyway",
		)
	}
	Sprites = append(Sprites, textExitSprites...)
	exitButton, err := CreateButton(64, 256, 512, 756)
	if err != nil {
		creationSuccess = false
		logger.LOG.Error().Err(err).Msg("")
	} else {
		mm.exitButton = exitButton
		mm.exitButton.OnPress = almostExitGame
		mm.exitButton.OnRelease = exitGame
		Sprites[1] = exitButton.Sprite
		AudioPlayers[0] = mm.exitButton.AudioPlayer
	}

	textSprites, ok := text.TextToSprites("Welcome to the Game!", 192, 384, 3, 20)
	if !ok {
		creationSuccess = false
		logger.LOG.Error().Msg("failed to make some of the title text sprites. trying to display anyway")
	}
	Sprites = append(Sprites, textSprites...)

	for _, sprite := range Sprites {
		if sprite == nil {
			logger.LOG.Error().Msg(
				"nil sprite encountered adding MainMenu sprites to draw queue (init)",
			)
		} else {
			sprites.GetDrawQueue().AddToQueue(weak.Make(sprite))
		}
	}

	return []scenes.GameObject{mm}, Sprites, AudioPlayers, creationSuccess
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
