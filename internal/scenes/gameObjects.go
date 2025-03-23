package scenes

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

// objects to be updated every game loop.
// Note that the Kill() can help define the lifetime of any created objects. This is in reference
// to the GC.
// Entire game lifetime => make the object a gameobject and attach it to globalScene
// Current scene lifetime => make the object a gameobject and attach it to the currentScene
// Some gameobject lifetime => make the object be called in the parent's kill function
type GameObject interface {
	InitInstance() ([]GameObject, []*sprites.Sprite, bool)
	Update()
	ShouldSkipUpdate() bool
	Kill()
}

// Inits provided game object and attaches resulting sprites, if any, and the game object
// itself to the current scene. If draw is true, it also adds the sprites to the draw queue.
func InitOnCurrentScene(gameObj GameObject, draw bool) {
	gameObjs, Sprites, ok := gameObj.InitInstance()
	if !ok {
		logger.LOG.Error().Msgf("Error creating game object: %v", gameObj)
		return
	}

	GetGlobalScene().currentScene.AddToSprites(Sprites...)
	GetGlobalScene().currentScene.AddToGameObjects(gameObjs...)
	if draw {
		for _, sprite := range Sprites {
			if sprite == nil {
				logger.LOG.Error().Msg("nil sprite encountered adding sprites to scene (init)")
			} else {
				sprites.GetDrawQueue().AddToQueue(weak.Make(sprite))
			}
		}
	}
}

// Inits provided game object and attaches resulting sprites, if any, and the game object
// itself to the global scene. This will preserve it between scene changes. The object will also
// be attached to the current scene, it possible.
// If draw is true, it also adds the sprites to the draw queue.
func InitOnGlobalScene(gameObj GameObject, draw bool) {
	gameObjs, Sprites, ok := gameObj.InitInstance()
	if !ok {
		logger.LOG.Error().Msgf("Error creating game object: %v", gameObj)
		return
	}

	GetGlobalScene().AddToSprites(Sprites...)
	GetGlobalScene().AddToGameObjects(gameObj)
	if GetGlobalScene().currentScene != nil {
		GetGlobalScene().currentScene.AddToSprites(Sprites...)
		GetGlobalScene().currentScene.AddToGameObjects(gameObjs...)
	}
	if draw {
		for _, sprite := range Sprites {
			if sprite == nil {
				logger.LOG.Error().Msg("nil sprite encountered adding sprites to scene (init)")
			} else {
				sprites.GetDrawQueue().AddToQueue(weak.Make(sprite))
			}
		}
	}
}

// Inits provided game object and attaches resulting sprites, if any, and the game object
// itself to the provided scene. If draw is true, it also adds the sprites to the draw queue.
func InitOnScene(scene *Scene, gameObj GameObject, draw bool) {
	gameObjs, Sprites, ok := gameObj.InitInstance()
	if !ok {
		logger.LOG.Error().Msgf("Error creating game object: %v", gameObj)
		return
	}

	scene.AddToSprites(Sprites...)
	scene.AddToGameObjects(gameObjs...)
	if draw {
		for _, sprite := range Sprites {
			if sprite == nil {
				logger.LOG.Error().Msg("nil sprite encountered adding sprites to scene (init)")
			} else {
				sprites.GetDrawQueue().AddToQueue(weak.Make(sprite))
			}
		}
	}
}
