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
}

var activeGlobalScene *globalScene
var once sync.Once

func GetGlobalScene() *globalScene {
	once.Do(func() { createGlobalScene() })
	return activeGlobalScene
}

func createGlobalScene() {
	// Call on the main thread
	activeGlobalScene = new(globalScene)
	activeGlobalScene.GlobalSprites = append(activeGlobalScene.GlobalSprites, cursor.GetCursor())
	// pc := new(gameObjects.PlayerCharacter)
	// pcSprites, ok := pc.InitInstance()
	// ... append(activeGameState.GlobalGameObjects, &pc)
	// ... append(activeGameState.GlobalSprites, ...pcSprite)
}

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

func isNextSceneRequested() bool {
	val, ok := gameState.GetCurrentGameState().GetFlagValue(gameState.NextScene)
	if !ok {
		return false
	}
	return val != 0
}

func wasCloseRequested() bool {
	val, ok := gameState.GetCurrentGameState().GetFlagValue(gameState.CloseRequested)
	if !ok {
		return false
	}
	gameState.GetCurrentGameState().SetFlagValue(gameState.CloseRequested, 0)
	return val == 1
}

func (gs *globalScene) useLoadingScene() bool {
	val, ok := gameState.GetCurrentGameState().GetFlagValue(gs.loadingSceneFlag)
	return ok && val != 0
}

func (gs *globalScene) addToScene(scene *Scene) {
	scene.Sprites = append(scene.Sprites, gs.GlobalSprites...)
	for _, gameObj := range gs.GlobalGameObjects {
		if gameObj == nil {
			continue
		}
		scene.GameObjects = append(gs.currentScene.GameObjects, *gameObj)
	}
}

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
		sprites.DrawDrawQueue()
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
		sprites.AddToDrawingQueue(weak.Make(sprite))
	}

	gs.currentScene = nextScene
}
