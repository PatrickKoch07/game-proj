package gameScenes

import (
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/myGame/gameUi"
	"github.com/PatrickKoch07/game-proj/internal/scenes"
)

func createTitleScene() *scenes.Scene {
	logger.LOG.Info().Msg("Making titleScene")

	titleScene := new(scenes.Scene)
	mm := gameUi.MainMenu{}
	scenes.InitOnScene(titleScene, scenes.GameObject(mm))
	return titleScene
}
