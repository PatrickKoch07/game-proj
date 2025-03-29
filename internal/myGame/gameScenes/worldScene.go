package gameScenes

import (
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/myGame/gameCharacters"
	"github.com/PatrickKoch07/game-proj/internal/scenes"
)

func createWorldScene() *scenes.Scene {
	logger.LOG.Info().Msg("Making worldScene")

	worldScene := new(scenes.Scene)
	player := new(gameCharacters.Player)
	scenes.InitOnGlobalScene(scenes.GameObject(player))
	block := new(gameCharacters.Block)
	scenes.InitOnGlobalScene(scenes.GameObject(block))
	return worldScene
}
