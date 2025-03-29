package scenes

import (
	"sync"

	"github.com/PatrickKoch07/game-proj/internal/audio"
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
	InitInstance() ([]GameObject, []*sprites.Sprite, []audio.Player, bool)
	Update()
	// If the gameobject is already loaded, but not yet 'ready' for play
	ShouldSkipUpdate() bool
	// The gameobject should clean up relavant data here too, like Clearing any audioplayers
	Kill()
	// If the gameobject has been killed and should be removed from other's memory (So can be GC'd)
	IsDead() bool
}

// Inits provided game object and attaches resulting sprites, if any, and the game object
// itself to the current scene.
func InitOnCurrentScene(gameObj GameObject) {
	gameObjs, Sprites, audioPlayers, ok := gameObj.InitInstance()
	if !ok {
		logger.LOG.Error().Msgf("Error creating game object: %v", gameObj)
		return
	}

	GetGlobalScene().currentScene.AddToSprites(Sprites...)
	GetGlobalScene().currentScene.AddToGameObjects(gameObjs...)
	GetGlobalScene().currentScene.AddToAudio(audioPlayers...)
}

// Inits provided game object and attaches resulting sprites, if any, and the game object
// itself to the global scene. This will preserve it between scene changes. The object will also
// be attached to the current scene, it possible.
func InitOnGlobalScene(gameObj GameObject) {
	gameObjs, Sprites, audioPlayers, ok := gameObj.InitInstance()
	if !ok {
		logger.LOG.Error().Msgf("Error creating game object: %v", gameObj)
		return
	}

	GetGlobalScene().AddToSprites(Sprites...)
	GetGlobalScene().AddToGameObjects(gameObjs...)
	GetGlobalScene().AddToAudio(audioPlayers...)
}

// Inits provided game object and attaches resulting sprites, if any, and the game object
// itself to the provided scene.
func InitOnScene(scene *Scene, gameObj GameObject) {
	gameObjs, Sprites, audioPlayers, ok := gameObj.InitInstance()
	if !ok {
		logger.LOG.Error().Msgf("Error creating game object: %v", gameObj)
		return
	}

	scene.AddToSprites(Sprites...)
	scene.AddToGameObjects(gameObjs...)
	scene.AddToAudio(audioPlayers...)
}

// should only be called from the main thread
func updateGameObjects(gameObjects []GameObject) []GameObject {
	var wg sync.WaitGroup
	maxIterInd := len(gameObjects)
	for i, gameObject := range gameObjects {
		// because we delete during a loop
		if i >= maxIterInd {
			break
		}

		if gameObject.IsDead() {
			// remove from slice
			gameObjects[i] = gameObjects[len(gameObjects)-1]
			gameObjects = gameObjects[:len(gameObjects)-1]
			i--
			continue
		}
		if gameObject.ShouldSkipUpdate() {
			continue
		}

		wg.Add(1)
		go func() { defer wg.Done(); gameObject.Update() }()
	}
	wg.Wait()
	return gameObjects
}
