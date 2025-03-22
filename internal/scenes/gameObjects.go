package scenes

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

type GameObject interface {
	InitInstance() ([]*sprites.Sprite, bool)
	Update()
	ShouldSkipUpdate() bool
}

// Inits provided game object and attaches resulting sprites, if any, and the game object
// itself to the current scene. If draw is true, it also adds the sprites to the draw queue.
func InitOnCurrentScene(gameObj GameObject, draw bool) {
	Sprites, ok := gameObj.InitInstance()
	if !ok {
		logger.LOG.Error().Msgf("Error creating game object: %v", gameObj)
		return
	}

	GetGlobalScene().currentScene.AddToSprites(Sprites...)
	GetGlobalScene().currentScene.AddToGameObjects(gameObj)
	if draw {
		for _, sprite := range Sprites {
			sprites.GetDrawQueue().AddToQueue(weak.Make(sprite))
		}
	}
}

// Inits provided game object and attaches resulting sprites, if any, and the game object
// itself to the global scene. This will preserve it between scene changes. The object will also
// be attached to the current scene, it possible.
// If draw is true, it also adds the sprites to the draw queue.
func InitOnGlobalScene(gameObj GameObject, draw bool) {
	Sprites, ok := gameObj.InitInstance()
	if !ok {
		logger.LOG.Error().Msgf("Error creating game object: %v", gameObj)
		return
	}

	GetGlobalScene().AddToSprites(Sprites...)
	GetGlobalScene().AddToGameObjects(&gameObj)
	if GetGlobalScene().currentScene != nil {
		GetGlobalScene().currentScene.AddToSprites(Sprites...)
		GetGlobalScene().currentScene.AddToGameObjects(gameObj)
	}
	if draw {
		for _, sprite := range Sprites {
			sprites.GetDrawQueue().AddToQueue(weak.Make(sprite))
		}
	}
}

// Inits provided game object and attaches resulting sprites, if any, and the game object
// itself to the provided scene. If draw is true, it also adds the sprites to the draw queue.
func InitOnScene(scene *Scene, gameObj GameObject, draw bool) {
	Sprites, ok := gameObj.InitInstance()
	if !ok {
		logger.LOG.Error().Msgf("Error creating game object: %v", gameObj)
		return
	}

	scene.AddToSprites(Sprites...)
	scene.AddToGameObjects(gameObj)
	if draw {
		for _, sprite := range Sprites {
			sprites.GetDrawQueue().AddToQueue(weak.Make(sprite))
		}
	}
}
