package scenes

import "github.com/PatrickKoch07/game-proj/internal/logger"

var worldScene *Scene

func GetWorldScene() *Scene {
	if worldScene == nil {
		createWorldScene()
	}
	return worldScene
}

func createWorldScene() {
	worldScene = new(Scene)
	worldScene.Init = func() { logger.LOG.Debug().Msg("Dummy log: Main world loaded") }
}
