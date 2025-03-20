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

var activeGameState *globalScene
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

func PopNextScene() *Scene {
	nextSceneGetterFunc, ok := sceneMap[nextSceneName]
	if !ok {
		logger.LOG.Error().Msgf("Bad scene name: %v", nextSceneName)
		return nil
	}
	nextSceneName = ""
	return nextSceneGetterFunc()
}

func IsNextSceneRequested() bool {
	return nextSceneName != ""
}

func GetGlobalScene() *globalScene {
	if activeGameState == nil {
		logger.LOG.Fatal().Msg("Game doesn't exist, state is nil")
	}
	return activeGameState
}

func GameStart() {
	// Call on the main thread
	activeGameState = new(globalScene)
	activeGameState.GlobalSprites = append(activeGameState.GlobalSprites, cursor.GetCursor())
	// pc := new(gameObjects.PlayerCharacter)
	// pcSprites, ok := pc.InitInstance()
	// ... append(activeGameState.GlobalGameObjects, &pc)
	// ... append(activeGameState.GlobalSprites, ...pcSprite)
	activeGameState.currentScene = GetTitleScene()
	activeGameState.currentScene.Init()
	addFromStateToScene(activeGameState.currentScene)
}

func addFromStateToScene(scene *Scene) {
	scene.Sprites = append(scene.Sprites, GetGlobalScene().GlobalSprites...)
	for _, gameObj := range GetGlobalScene().GlobalGameObjects {
		if gameObj == nil {
			continue
		}
		scene.GameObjects = append(scene.GameObjects, *gameObj)
	}
}

func Update() {
	// Call on the main thread
	if ui.WasCloseRequested() {
		glfw.GetCurrentContext().SetShouldClose(true)
		return
	}
	if IsNextSceneRequested() {
		SwitchScene()
	}
	UpdateSceneGameObjects(GetGlobalScene().currentScene)
}

func SwitchScene() {
	nextScene := PopNextScene()
	if nextScene == nil {
		logger.LOG.Error().Msg("Bad scene name give to switch to. Ignoring switch.")
		return
	}
	// block below draws the loading screen
	StopDrawingScene(GetGlobalScene().currentScene)
	logger.LOG.Debug().Msg("Drawing loading screen")
	GetLoadingScene().Init()
	// clear previous rendering
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	// draw
	sprites.DrawDrawQueue()
	glfw.GetCurrentContext().SwapBuffers()

	// gameState specific logic goes here
	//

	// load next scene
	addFromStateToScene(nextScene)
	nextScene.Init()
	logger.LOG.Debug().Msg("Next scene loaded, removing unused graphics objects")
	UnloadUncommonGraphicObjs(GetGlobalScene().currentScene, nextScene)
	for _, sprite := range GetGlobalScene().GlobalSprites {
		sprites.AddToDrawingQueue(weak.Make(sprite))
	}
	// dummy line to let me see the loading screen
	time.Sleep(2 * time.Second)
	StopDrawingScene(GetLoadingScene())

	GetGlobalScene().currentScene = nextScene
}
