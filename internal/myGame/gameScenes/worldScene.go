package gameScenes

import (
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/scenes"
)

func createWorldScene() *scenes.Scene {
	worldScene := new(scenes.Scene)
	logger.LOG.Debug().Msg("Dummy log: Main world loaded")
	return worldScene
}
