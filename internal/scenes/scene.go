package scenes

import (
	"slices"
	"sync"

	"github.com/PatrickKoch07/game-proj/internal/audio"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

// a context to bind sprites, game objects and audio files, in game terms this is like a level
type Scene struct {
	// The '(maybe)' below refer to if the object is in the globalScene,
	// they shouldn't be affected.
	// On switch, what graphics objects to stop drawing & which to (maybe) delete between scenes
	Sprites []*sprites.Sprite
	// What objects to update each frame and (maybe) kill between scenes
	GameObjects []GameObject
	// What audio objects to stop playing and (maybe) close between scenes.
	AudioPlayers []*audio.Player
	mu           sync.Mutex
}

// (should be) called in the main thread
func unloadUncommonGraphicObjs(current *Scene, next *Scene, globalSprites []*sprites.Sprite) {
	nextShaders := make(map[uint32]struct{})
	nextTextures := make(map[uint32]struct{})
	nextVAOs := make(map[uint32]struct{})
	for _, sprite := range next.Sprites {
		nextShaders[sprite.GetShaderId()] = struct{}{}
		nextTextures[sprite.GetTextureId()] = struct{}{}
		nextVAOs[sprite.GetVAO()] = struct{}{}
	}
	for _, sprite := range globalSprites {
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

// kills gameobjects, clears sprites, and clears audio. any race-y calls on objects killing them
// should be okay because they have the same outcome ... ...
func killScene(s *Scene) {
	go clearSceneSprites(s)
	go clearSceneAudio(s)
	go killSceneGameObjects(s)
}

func clearSceneSprites(s *Scene) {
	for _, sprite := range s.Sprites {
		if sprite == nil {
			continue
		}
		err := sprite.Clear()
		if err != nil {
			logger.LOG.Warn().Err(err).Msg("Trying to continue anyway")
		}
	}
}

func clearSceneAudio(s *Scene) {
	for _, audioPlayer := range s.AudioPlayers {
		if audioPlayer == nil {
			logger.LOG.Warn().Msg("Nil audioPlayer")
			continue
		}
		err := (*audioPlayer).Clear()
		if err != nil {
			logger.LOG.Warn().Err(err).Msg("Trying to continue anyway")
		}
	}
}

func killSceneGameObjects(s *Scene) {
	for _, gameObj := range s.GameObjects {
		if !gameObj.IsDead() {
			gameObj.Kill()
		}
	}
}

// thread safe by locking
func (s *Scene) AddToSprites(newSprites ...*sprites.Sprite) {
	s.mu.Lock()
	s.Sprites = append(s.Sprites, newSprites...)
	if len(s.Sprites) > largeNumberOfSprites {
		s.Sprites = slices.DeleteFunc(
			s.Sprites,
			func(heldSprite *sprites.Sprite) bool {
				return heldSprite.IsNil()
			},
		)
	}
	s.mu.Unlock()
}

// thread safe by locking
func (s *Scene) RemoveFromSprites(sprite *sprites.Sprite) {
	s.mu.Lock()
	var index int = -1
	for ind, val := range s.Sprites {
		if val == sprite {
			index = ind
			break
		}
	}
	if index == -1 {
		return
	}
	s.Sprites[index] = s.Sprites[len(s.Sprites)-1]
	s.Sprites = s.Sprites[:len(s.Sprites)-1]
	s.mu.Unlock()
}

// thread safe by locking
func (s *Scene) AddToGameObjects(gameObjs ...GameObject) {
	s.mu.Lock()
	s.GameObjects = append(s.GameObjects, gameObjs...)
	s.mu.Unlock()
}

// thread safe by locking
func (s *Scene) RemoveFromGameObjects(gameObj GameObject) {
	s.mu.Lock()
	var index int = -1
	for ind, val := range s.GameObjects {
		if val == gameObj {
			index = ind
			break
		}
	}
	if index == -1 {
		return
	}
	s.GameObjects[index] = s.GameObjects[len(s.GameObjects)-1]
	s.GameObjects = s.GameObjects[:len(s.GameObjects)-1]
	s.mu.Unlock()
}

func (s *Scene) AddToAudio(audioPlayers ...*audio.Player) {
	s.mu.Lock()
	s.AudioPlayers = append(s.AudioPlayers, audioPlayers...)
	s.mu.Unlock()
}

func (s *Scene) RemoveFromAudio(audioPlayer *audio.Player) {
	s.mu.Lock()
	var index int = -1
	for ind, val := range s.AudioPlayers {
		if &val == &audioPlayer {
			index = ind
			break
		}
	}
	if index == -1 {
		return
	}
	s.AudioPlayers[index] = s.AudioPlayers[len(s.AudioPlayers)-1]
	s.AudioPlayers = s.AudioPlayers[:len(s.AudioPlayers)-1]
	s.mu.Unlock()
}
