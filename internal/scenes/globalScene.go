package scenes

import (
	"sync"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/cursor"
	"github.com/PatrickKoch07/game-proj/internal/gameState"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type globalScene struct {
	// object instances to update not related to any scene in particular (ex. player character)
	GlobalGameObjects []*GameObject
	// sprites to keep on display (ex. the cursor or some UI)
	GlobalSprites    []*sprites.Sprite
	loadingSceneFlag gameState.Flag
	sceneMap         map[gameState.Flag]func() *Scene
	currentScene     *Scene
	mu               sync.Mutex
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
	activeGlobalScene.addToScene(activeGlobalScene.currentScene)
	gs.loadingSceneFlag = loadingScene
}

// thread safe by locking
func (gs *globalScene) AddToSprites(sprites ...*sprites.Sprite) {
	gs.mu.Lock()
	gs.GlobalSprites = append(gs.GlobalSprites, sprites...)
	gs.mu.Unlock()
}

// thread safe by locking
func (gs *globalScene) RemoveFromSprites(sprite *sprites.Sprite) {
	gs.mu.Lock()
	var index int = -1
	for ind, val := range gs.GlobalSprites {
		if val == sprite {
			index = ind
			break
		}
	}
	if index == -1 {
		return
	}
	gs.GlobalSprites[index] = gs.GlobalSprites[len(gs.GlobalSprites)-1]
	gs.GlobalSprites = gs.GlobalSprites[:len(gs.GlobalSprites)-1]
	gs.mu.Unlock()
}

// thread safe by locking
func (gs *globalScene) RemoveFromGameObjects(gameObj *GameObject) {
	gs.mu.Lock()
	var index int = -1
	for ind, val := range gs.GlobalGameObjects {
		if val == gameObj {
			index = ind
			break
		}
	}
	if index == -1 {
		return
	}
	gs.GlobalGameObjects[index] = gs.GlobalGameObjects[len(gs.GlobalGameObjects)-1]
	gs.GlobalGameObjects = gs.GlobalGameObjects[:len(gs.GlobalGameObjects)-1]
	gs.mu.Unlock()
}

// thread safe by locking
func (gs *globalScene) AddToGameObjects(gameObjs ...*GameObject) {
	gs.mu.Lock()
	gs.GlobalGameObjects = append(gs.GlobalGameObjects, gameObjs...)
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
func (gs *globalScene) addToScene(scene *Scene) {
	scene.Sprites = append(scene.Sprites, gs.GlobalSprites...)
	for _, gameObj := range gs.GlobalGameObjects {
		if gameObj == nil {
			continue
		}
		scene.GameObjects = append(gs.currentScene.GameObjects, *gameObj)
	}
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
	updateSceneGameObjects(gs.currentScene)
}

// should only be called in the main thread
func (gs *globalScene) switchScene() {
	nextSceneFunc, ok := gs.popNextScene()
	if !ok {
		logger.LOG.Error().Msg("Ignoring scene switch.")
		return
	}

	// block below draws the loading screen
	stopDrawingScene(gs.currentScene)

	if gs.useLoadingScene() {
		logger.LOG.Debug().Msg("Drawing loading screen")
		loadingScene := gs.sceneMap[gs.loadingSceneFlag]()
		defer stopDrawingScene(loadingScene)
		// clear previous rendering
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Clear(gl.DEPTH_BUFFER_BIT)
		// draw
		sprites.GetDrawQueue().DrawDrawQueue()
		glfw.GetCurrentContext().SwapBuffers()
	}

	// gameState specific logic goes here
	// => load gameobject info of scene, if exists
	//

	// create next scene
	nextScene := nextSceneFunc()
	gs.addToScene(nextScene)
	logger.LOG.Debug().Msg("Next scene loaded, removing unused graphics objects")
	unloadUncommonGraphicObjs(gs.currentScene, nextScene)
	for _, sprite := range gs.GlobalSprites {
		sprites.GetDrawQueue().AddToDrawingQueue(weak.Make(sprite))
	}

	gs.currentScene = nextScene
}
