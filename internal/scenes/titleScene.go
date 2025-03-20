package scenes

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
	"github.com/PatrickKoch07/game-proj/internal/ui"
)

var titleScene *Scene

func GetTitleScene() *Scene {
	if titleScene == nil {
		createTitleScene()
	}
	return titleScene
}

func createTitleScene() {
	titleScene = new(Scene)
	titleScene.Init = initTitleScene
}

func initTitleScene() {
	// should be a better solution than reallllly remembering to do this before the init
	mm := ui.MainMenu{SwitchSceneFunc: flagSceneSwitch}
	buttonSprites, ok := mm.InitInstance()
	if !ok {
		logger.LOG.Error().Msg("Issue initializing main menu")
	} else {
		for _, sprite := range buttonSprites {
			sprites.AddToDrawingQueue(weak.Make(sprite))
		}
		GetTitleScene().GameObjects = append(GetTitleScene().GameObjects, mm)
		GetTitleScene().Sprites = append(GetTitleScene().Sprites, buttonSprites...)
	}
}

func flagSceneSwitch() {
	logger.LOG.Debug().Msg("Starting switch scene process")
	nextSceneName = "worldScene"
}
