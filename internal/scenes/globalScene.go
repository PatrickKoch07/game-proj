package scenes

import (
	"time"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/cursor"
	"github.com/PatrickKoch07/game-proj/internal/gameObjects"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
	"github.com/PatrickKoch07/game-proj/internal/ui"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// should be fixed by gamestate TODO
var nextSceneName string

type globalScene struct {
	// object instances to update not related to any scene in particular (ex. player character)
	GlobalGameObjects []*gameObjects.GameObject
	// sprites to keep on display (ex. the cursor or some UI)
	GlobalSprites []*sprites.Sprite
	currentScene  *Scene
}

func init() {
	nextSceneName = ""
}

func popNextScene() *Scene {
	defer func() { nextSceneName = "" }()

	switch nextSceneName {
	case "titleScene":
		return createTitleScene()
	case "worldScene":
		return createWorldScene()
	default:
		logger.LOG.Error().Msgf("Bad scene name: %v", nextSceneName)
		return nil
	}
}

func isNextSceneRequested() bool {
	return nextSceneName != ""
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
	activeGameState.addToCurrentScene()
	return activeGameState
}

func (gs *globalScene) addToCurrentScene() {
	gs.currentScene.Sprites = append(gs.currentScene.Sprites, gs.GlobalSprites...)
	for _, gameObj := range gs.GlobalGameObjects {
		if gameObj == nil {
			continue
		}
		gs.currentScene.GameObjects = append(gs.currentScene.GameObjects, *gameObj)
	}
}

func (gs *globalScene) Update() {
	// Call on the main thread
	if ui.WasCloseRequested() {
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
		logger.LOG.Error().Msg("Bad scene name give to switch to. Ignoring switch.")
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
	//

	// load next scene
	gs.addToCurrentScene()
	nextScene.Init(nextScene)
	logger.LOG.Debug().Msg("Next scene loaded, removing unused graphics objects")
	UnloadUncommonGraphicObjs(gs.currentScene, nextScene)
	for _, sprite := range gs.GlobalSprites {
		sprites.AddToDrawingQueue(weak.Make(sprite))
	}
	// dummy line to let me see the loading screen
	time.Sleep(2 * time.Second)
	StopDrawingScene(loadingScene)

	gs.currentScene = nextScene
}
