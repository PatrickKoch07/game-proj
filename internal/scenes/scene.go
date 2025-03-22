package scenes

import (
	"sync"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

type Scene struct {
	// On switch, what graphics objects to potentially stop drawing & which to switch out
	Sprites []*sprites.Sprite
	// What objects to update
	GameObjects []GameObject
}

func unloadUncommonGraphicObjs(current *Scene, next *Scene) {
	nextShaders := make(map[uint32]struct{})
	nextTextures := make(map[uint32]struct{})
	nextVAOs := make(map[uint32]struct{})
	for _, sprite := range next.Sprites {
		nextShaders[sprite.GetShaderId()] = struct{}{}
		nextTextures[sprite.GetTextureId()] = struct{}{}
		nextVAOs[sprite.GetVAO()] = struct{}{}
	}

	// if not in the next scene, delete
	for _, sprite := range current.Sprites {
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
	for _, sprite := range s.Sprites {
		if sprite == nil {
			continue
		}
		ok := sprites.RemoveFromDrawingQueue(weak.Make(sprite))
		if !ok {
			logger.LOG.Warn().Msg("Issue with removing object from drawing queue. Trying to continue")
		}
	}
}

func updateSceneGameObjects(currentScene *Scene) {
	// called in main game loop
	var wg sync.WaitGroup
	for _, gameObject := range currentScene.GameObjects {
		if gameObject.ShouldSkipUpdate() {
			continue
		}
		wg.Add(1)
		go func() { defer wg.Done(); gameObject.Update() }()
	}
	wg.Wait()
}
