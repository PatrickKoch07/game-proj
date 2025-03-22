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

	GetGlobalScene().currentScene.Sprites = append(
		GetGlobalScene().currentScene.Sprites, Sprites...,
	)
	GetGlobalScene().currentScene.GameObjects = append(
		GetGlobalScene().currentScene.GameObjects, gameObj,
	)
	if draw {
		for _, sprite := range Sprites {
			sprites.AddToDrawingQueue(weak.Make(sprite))
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
	GetGlobalScene().GlobalSprites = append(GetGlobalScene().GlobalSprites, Sprites...)
	GetGlobalScene().GlobalGameObjects = append(GetGlobalScene().GlobalGameObjects, &gameObj)
	if GetGlobalScene().currentScene != nil {
		GetGlobalScene().currentScene.Sprites = append(
			GetGlobalScene().currentScene.Sprites, Sprites...,
		)
		GetGlobalScene().currentScene.GameObjects = append(
			GetGlobalScene().currentScene.GameObjects, gameObj,
		)
	}
	if draw {
		for _, sprite := range Sprites {
			sprites.AddToDrawingQueue(weak.Make(sprite))
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

	scene.Sprites = append(
		scene.Sprites, Sprites...,
	)
	scene.GameObjects = append(
		scene.GameObjects, gameObj,
	)
	if draw {
		for _, sprite := range Sprites {
			sprites.AddToDrawingQueue(weak.Make(sprite))
		}
	}
}
