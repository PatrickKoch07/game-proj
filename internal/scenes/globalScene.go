package scenes

import (
	"time"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/cursor"
	"github.com/PatrickKoch07/game-proj/internal/gameObjects"
	"github.com/PatrickKoch07/game-proj/internal/gameState"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type globalScene struct {
	// object instances to update not related to any scene in particular (ex. player character)
	GlobalGameObjects []*gameObjects.GameObject
	// sprites to keep on display (ex. the cursor or some UI)
	GlobalSprites []*sprites.Sprite
	currentScene  *Scene
}

func popNextScene() *Scene {
	val, ok := gameState.GetCurrentGameState().GetFlagValue(gameState.NextScene)
	if !ok || val == 0 {
		return nil
	}

	defer func() { gameState.GetCurrentGameState().SetFlagValue(gameState.NextScene, 0) }()

	switch val {
	case int(gameState.TitleScene):
		return createTitleScene()
	case int(gameState.WorldScene):
		return createWorldScene()
	default:
		logger.LOG.Error().Msgf("Bad scene value: %v", val)
		return nil
	}
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

func CreateGlobalScene() *globalScene {
	// Call on the main thread
	activeGameState := new(globalScene)
	activeGameState.GlobalSprites = append(activeGameState.GlobalSprites, cursor.GetCursor())
	// pc := new(gameObjects.PlayerCharacter)
	// pcSprites, ok := pc.InitInstance()
	// ... append(activeGameState.GlobalGameObjects, &pc)
	// ... append(activeGameState.GlobalSprites, ...pcSprite)
	activeGameState.currentScene = createTitleScene()
	activeGameState.currentScene.Init(activeGameState.currentScene)
	activeGameState.addToScene(activeGameState.currentScene)
	return activeGameState
}

func (gs *globalScene) addToScene(scene *Scene) {
	scene.Sprites = append(scene.Sprites, gs.GlobalSprites...)
	for _, gameObj := range gs.GlobalGameObjects {
		if gameObj == nil {
			continue
		}
		gs.currentScene.GameObjects = append(gs.currentScene.GameObjects, *gameObj)
	}
}

func (gs *globalScene) Update() {
	// Call on the main thread
	if wasCloseRequested() {
		glfw.GetCurrentContext().SetShouldClose(true)
		return
	}
	if isNextSceneRequested() {
		gs.SwitchScene()
	}
	UpdateSceneGameObjects(gs.currentScene)
}

func (gs *globalScene) SwitchScene() {
	nextScene := popNextScene()
	if nextScene == nil {
		logger.LOG.Error().Msg("Ignoring scene switch.")
		return
	}
	// block below draws the loading screen
	StopDrawingScene(gs.currentScene)
	logger.LOG.Debug().Msg("Drawing loading screen")
	loadingScene := createLoadingScene()
	loadingScene.Init(loadingScene)
	// clear previous rendering
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	// draw
	sprites.DrawDrawQueue()
	glfw.GetCurrentContext().SwapBuffers()

	// gameState specific logic goes here
	// => load gameobject info of scene, if exists
	//

	// create next scene
	gs.addToScene(nextScene)
	nextScene.Init(nextScene)
	logger.LOG.Debug().Msg("Next scene loaded, removing unused graphics objects")
	UnloadUncommonGraphicObjs(gs.currentScene, nextScene)
	for _, sprite := range gs.GlobalSprites {
		sprites.AddToDrawingQueue(weak.Make(sprite))
	}
	// try to load previous scene info
	// dummy line to let me see the loading screen
	time.Sleep(2 * time.Second)
	StopDrawingScene(loadingScene)

	gs.currentScene = nextScene
}
