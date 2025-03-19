package scenes

import "github.com/PatrickKoch07/game-proj/internal/logger"

var worldScene *Scene

func GetWorldScene() *Scene {
	if worldScene == nil {
		createLoadingScene()
	}
	return worldScene
}

func createWorldScene() {
	worldScene.Init = func() { logger.LOG.Debug().Msg("Main world loaded") }
}
