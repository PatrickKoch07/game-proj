package scenes

import "github.com/PatrickKoch07/game-proj/internal/logger"

func createWorldScene() *Scene {
	worldScene := new(Scene)
	worldScene.Init = func(_ *Scene) { logger.LOG.Debug().Msg("Dummy log: Main world loaded") }
	return worldScene
}
