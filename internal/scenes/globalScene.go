package scenes

import (
	"slices"
	"sync"
	"time"

	"github.com/PatrickKoch07/game-proj/internal/audio"
	"github.com/PatrickKoch07/game-proj/internal/cursor"
	"github.com/PatrickKoch07/game-proj/internal/gameState"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const largeNumberOfSprites int = 100
const largeNumberOfAudioPlayers int = 20

// context to bind sprites, game objects and audio to. Separate from scene as this is meant to last
// for the whole duration of the game. as such, handles the switching between scenes
type globalScene struct {
	// object instances to update not related to any scene in particular (ex. player character).
	GlobalGameObjects []GameObject
	// sprites to keep on display (ex. the cursor or some UI)
	// Even if these sprites are kept referenced by the gameobjects, its nice to directly know
	// which sprites are still used when scenes change so their shader info doesn't delete.
	GlobalSprites []*sprites.Sprite
	// audio objects to keep on (ex. audio related to the gameObjects)
	// Likewise to sprites, audio kept here will still be playing/not be reloaded between scene
	// switching.
	GlobalAudioPlayers []audio.Player
	// technically this could be user defined, but with how many loading screens there are in
	// games these days, I added it to the core game engine structure
	loadingSceneFlag gameState.Flag
	// To be loaded from the main function on game start. whatever is fed in should be user defined
	sceneMap map[gameState.Flag]func() *Scene

	currentScene *Scene
	mu           sync.Mutex
}

var activeGlobalScene *globalScene
var once sync.Once

func GetGlobalScene() *globalScene {
	once.Do(func() { createGlobalScene() })
	return activeGlobalScene
}

func createGlobalScene() {
	activeGlobalScene = new(globalScene)
	activeGlobalScene.GlobalSprites = append(activeGlobalScene.GlobalSprites, cursor.GetCursor())
}

// should only be called in the main thread
func (gs *globalScene) InitializeGlobalScene(
	sceneMap map[gameState.Flag]func() *Scene,
	firstScene gameState.Flag,
	loadingScene gameState.Flag,
) {
	gs.sceneMap = make(map[gameState.Flag]func() *Scene)
	gs.sceneMap = sceneMap
	gs.currentScene = sceneMap[firstScene]()
	gs.loadingSceneFlag = loadingScene
}

// thread safe by locking
func (gs *globalScene) AddToSprites(newSprites ...*sprites.Sprite) {
	gs.mu.Lock()
	gs.GlobalSprites = append(gs.GlobalSprites, newSprites...)
	if len(gs.GlobalSprites) > largeNumberOfSprites {
		gs.GlobalSprites = slices.DeleteFunc(
			gs.GlobalSprites,
			func(heldSprite *sprites.Sprite) bool {
				return heldSprite.IsNil()
			},
		)
	}
	gs.mu.Unlock()
}

// thread safe by locking
func (gs *globalScene) RemoveFromSprites(sprite *sprites.Sprite) {
	gs.mu.Lock()
	// var index int = -1
	// for ind, val := range gs.GlobalSprites {
	// 	if val == sprite {
	// 		index = ind
	// 		break
	// 	}
	// }
	// if index == -1 {
	// 	return
	// }
	// gs.GlobalSprites[index] = gs.GlobalSprites[len(gs.GlobalSprites)-1]
	// gs.GlobalSprites = gs.GlobalSprites[:len(gs.GlobalSprites)-1]

	gs.GlobalSprites = slices.DeleteFunc(
		gs.GlobalSprites,
		func(heldSprite *sprites.Sprite) bool {
			return &heldSprite == &sprite
		},
	)

	gs.mu.Unlock()
}

// thread safe by locking
func (gs *globalScene) AddToGameObjects(gameObjs ...GameObject) {
	gs.mu.Lock()
	gs.GlobalGameObjects = append(gs.GlobalGameObjects, gameObjs...)
	gs.mu.Unlock()
}

// thread safe by locking
func (gs *globalScene) RemoveFromGameObjects(gameObj GameObject) {
	gs.mu.Lock()
	// var index int = -1
	// for ind, val := range gs.GlobalGameObjects {
	// 	if val == gameObj {
	// 		index = ind
	// 		break
	// 	}
	// }
	// if index == -1 {
	// 	return
	// }
	// gs.GlobalGameObjects[index] = gs.GlobalGameObjects[len(gs.GlobalGameObjects)-1]
	// gs.GlobalGameObjects = gs.GlobalGameObjects[:len(gs.GlobalGameObjects)-1]

	gs.GlobalGameObjects = slices.DeleteFunc(
		gs.GlobalGameObjects,
		func(heldGameObj GameObject) bool {
			return heldGameObj == gameObj
		},
	)

	gs.mu.Unlock()
}

func (gs *globalScene) AddToAudio(audioPlayers ...audio.Player) {
	gs.mu.Lock()
	gs.GlobalAudioPlayers = append(gs.GlobalAudioPlayers, audioPlayers...)
	if len(gs.GlobalAudioPlayers) > largeNumberOfAudioPlayers {
		gs.GlobalAudioPlayers = slices.DeleteFunc(
			gs.GlobalAudioPlayers,
			func(heldAudioPlayer audio.Player) bool {
				return heldAudioPlayer.IsNil()
			},
		)
	}
	gs.mu.Unlock()
}

func (gs *globalScene) RemoveFromAudio(audioPlayer audio.Player) {
	gs.mu.Lock()
	gs.GlobalAudioPlayers = slices.DeleteFunc(
		gs.GlobalAudioPlayers,
		func(heldAudioPlayer audio.Player) bool {
			return heldAudioPlayer == audioPlayer
		},
	)
	gs.mu.Unlock()
}

// should only be called in the main thread
func (gs *globalScene) popNextScene() (func() *Scene, bool) {
	default_func := func() *Scene { return new(Scene) }

	val, ok := gameState.GetCurrentGameState().GetFlagValue(gameState.NextScene)
	defer func() { gameState.GetCurrentGameState().SetFlagValue(gameState.NextScene, 0) }()
	if !ok || val == 0 {
		return default_func, false
	}

	nextSceneFunc, ok := gs.sceneMap[gameState.Flag(val)]
	if !ok {
		logger.LOG.Error().Msgf("Bad scene value: %v", val)
		return default_func, false
	}

	return nextSceneFunc, true
}

// should only be called in the main thread
func isNextSceneRequested() bool {
	val, ok := gameState.GetCurrentGameState().GetFlagValue(gameState.NextScene)
	if !ok {
		return false
	}
	return val != 0
}

// should only be called in the main thread
func wasCloseRequested() bool {
	val, ok := gameState.GetCurrentGameState().GetFlagValue(gameState.CloseRequested)
	if !ok {
		return false
	}
	gameState.GetCurrentGameState().SetFlagValue(gameState.CloseRequested, 0)
	return val == 1
}

// should only be called in the main thread
func (gs *globalScene) useLoadingScene() bool {
	val, ok := gameState.GetCurrentGameState().GetFlagValue(gs.loadingSceneFlag)
	return ok && val != 0
}

// should only be called in the main thread
func (gs *globalScene) Update() {
	// Call on the main thread
	if wasCloseRequested() {
		glfw.GetCurrentContext().SetShouldClose(true)
		return
	}
	if isNextSceneRequested() {
		gs.switchScene()
	}

	gs.currentScene.GameObjects = updateGameObjects(gs.currentScene.GameObjects)
	gs.GlobalGameObjects = updateGameObjects(gs.GlobalGameObjects)
}

// should only be called in the main thread
func (gs *globalScene) switchScene() {
	nextSceneFunc, ok := gs.popNextScene()
	if !ok {
		logger.LOG.Error().Msg("Ignoring scene switch.")
		return
	}

	Kill(gs.currentScene)

	if gs.useLoadingScene() {
		logger.LOG.Debug().Msg("Drawing loading screen")
		loadingScene := gs.sceneMap[gs.loadingSceneFlag]()
		defer Kill(loadingScene)
		// clear previous rendering
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Clear(gl.DEPTH_BUFFER_BIT)
		// draw
		sprites.GetDrawQueue().Draw()
		glfw.GetCurrentContext().SwapBuffers()
		// I added this so I can see some of the loading screen at least
		time.Sleep(1 * time.Second)
	}

	// gameState specific logic goes here
	// ex. load info of scene, if exists
	//

	// clean up resources
	gs.GlobalAudioPlayers = slices.DeleteFunc(
		gs.GlobalAudioPlayers,
		func(audioPlayer audio.Player) bool {
			return audioPlayer == nil || audioPlayer.IsNil()
		},
	)
	gs.GlobalSprites = slices.DeleteFunc(
		gs.GlobalSprites,
		func(heldSprite *sprites.Sprite) bool {
			return heldSprite == nil || heldSprite.IsNil()
		},
	)
	// create next scene
	nextScene := nextSceneFunc()
	logger.LOG.Debug().Msg("Next scene loaded, removing unused graphics objects")
	unloadUncommonGraphicObjs(gs.currentScene, nextScene, gs.GlobalSprites)

	gs.currentScene = nextScene
}

func (gs *globalScene) Kill() {
	go gs.clearSprites()
	go gs.clearAudio()
	go gs.killGameObjects()
	go Kill(gs.currentScene)
}

func (gs *globalScene) clearSprites() {
	for _, sprite := range gs.GlobalSprites {
		if sprite == nil {
			continue
		}
		err := sprite.Clear()
		if err != nil {
			logger.LOG.Warn().Err(err).Msg("Trying to continue anyway")
		}
	}
}

func (gs *globalScene) clearAudio() {
	for _, audioPlayer := range gs.GlobalAudioPlayers {
		if audioPlayer == nil {
			logger.LOG.Warn().Msg("Nil audioPlayer")
			continue
		}
		err := audioPlayer.Clear()
		if err != nil {
			logger.LOG.Warn().Err(err).Msg("Trying to continue anyway")
		}
	}
}

func (gs *globalScene) killGameObjects() {
	for _, gameObj := range gs.GlobalGameObjects {
		if !gameObj.IsDead() {
			gameObj.Kill()
		}
	}
}
