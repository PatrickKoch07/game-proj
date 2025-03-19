package scenes

import (
	"time"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/gameObjects"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var nextScene *Scene

type Scene struct {
	sprites     []*sprites.Sprite
	gameObjects []gameObjects.GameObject
	Init        func()
}

func IsNextSceneRequested() bool {
	return nextScene != nil
}

func SetNextScene(next *Scene) {
	nextScene = next
}

func unloadUncommonGraphicObjs(current *Scene, next *Scene) {
	nextShaders := make(map[uint32]struct{})
	nextTextures := make(map[uint32]struct{})
	nextVAOs := make(map[uint32]struct{})
	for _, sprite := range next.sprites {
		nextShaders[sprite.GetShaderId()] = struct{}{}
		nextTextures[sprite.GetTextureId()] = struct{}{}
		nextVAOs[sprite.GetVAO()] = struct{}{}
	}

	// if not in the next scene, delete
	for _, sprite := range current.sprites {
		_, ok := nextShaders[sprite.GetShaderId()]
		if !ok {
			sprites.DeleteShaderById(sprite.GetShaderId())
			logger.LOG.Debug().Msgf("deleted shader %v", sprite.GetShaderId())
		}
		_, ok = nextTextures[sprite.GetTextureId()]
		if !ok {
			sprites.DeleteTextureById(sprite.GetTextureId())
			logger.LOG.Debug().Msgf("deleted texture %v", sprite.GetTextureId())
		}
		_, ok = nextVAOs[sprite.GetVAO()]
		if !ok {
			sprites.DeleteVAOById(sprite.GetVAO())
			logger.LOG.Debug().Msgf("deleted vao %v", sprite.GetVAO())
		}
	}
}

func stopDrawingScene(s *Scene) {
	for _, sprite := range s.sprites {
		if sprite == nil {
			continue
		}
		ok := sprites.RemoveFromDrawingQueue(weak.Make(sprite))
		if !ok {
			logger.LOG.Warn().Msg("Issue with removing object from drawing queue. Trying to continue")
		}
	}
}

func SwitchScene(currentScene *Scene, window *glfw.Window) *Scene {
	// MUST be called in main game loop/thread

	stopDrawingScene(currentScene)
	loadingScene.Init()
	// clear previous rendering
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	// draw
	sprites.DrawDrawQueue()
	window.SwapBuffers()
	unloadUncommonGraphicObjs(currentScene, nextScene)

	nextScene.Init()
	// dummy to let me see the loading screen
	time.Sleep(10 * time.Second)
	stopDrawingScene(GetLoadingScene())

	currentScene = nextScene
	SetNextScene(nil)
	return currentScene
}

func UpdateSceneGameObjects(currentScene *Scene) {
	// called in main game loop
	for _, gameObject := range currentScene.gameObjects {
		if gameObject.ShouldSkipUpdate() {
			continue
		}
		go gameObject.Update()
	}
}
