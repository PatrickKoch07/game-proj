package scenes

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
	"github.com/PatrickKoch07/game-proj/internal/ui"
)

func createTitleScene() *Scene {
	titleScene := new(Scene)
	titleScene.Init = initTitleScene
	return titleScene
}

func initTitleScene(titleScene *Scene) {
	// should be a better solution than reallllly remembering to do this before the init
	mm := ui.MainMenu{}
	buttonSprites, ok := mm.InitInstance()
	if !ok {
		logger.LOG.Error().Msg("Issue initializing main menu")
	} else {
		for _, sprite := range buttonSprites {
			sprites.AddToDrawingQueue(weak.Make(sprite))
		}
		titleScene.GameObjects = append(titleScene.GameObjects, mm)
		titleScene.Sprites = append(titleScene.Sprites, buttonSprites...)
	}
}
