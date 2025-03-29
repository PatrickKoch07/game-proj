package sprites

import (
	"container/list"
	"sync"
	"sync/atomic"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type Sprite struct {
	shaderId uint32
	Tex      texture
	vao      uint32
	// Where the origin of the object is on the screen (top left is 0.0, 0.0)
	ScreenCenter ScreenCoords
	// from 0.0 to 1.0, Where the origin of the object is on the sprite (top left is 0.0, 0.0)
	SpriteCenter SpriteCoords

	// Sprites cannot be deleted in isolation because the shaderId, textureId, or VAO might be used
	// by some other object. So this is marked for lazy deletion (do not draw), to be deleted
	// from the GPU and 'cache' system in shader.go when its okay to do the expensive checking.
	lazyDeletionMark atomic.Bool
}

func (s *Sprite) GetShaderId() uint32 {
	return s.shaderId
}

func (s *Sprite) GetTextureId() uint32 {
	return s.Tex.textureId
}

func (s *Sprite) GetVAO() uint32 {
	return s.vao
}

func (s *Sprite) Clear() error {
	GetDrawQueue().RemoveFromQueue(weak.Make(s))
	s.lazyDeletionMark.Store(true)
	return nil
}

func (s *Sprite) IsNil() bool {
	return s.lazyDeletionMark.Load()
}

type ScreenCoords struct {
	X float32
	Y float32
}

// X and Y must be between 0.0 and 1.0
type SpriteCoords struct {
	X float32
	Y float32
}

type SpriteInitParams struct {
	ShaderRelPaths ShaderFiles
	TextureRelPath string
	// TextureCoords must be: Bottom left, top left, top right, bottom left, top right, bottom right
	TextureCoords [12]float32
	// default position in screen: top left (0, 0)
	ScreenCenter ScreenCoords
	// default origin of sprite: upper left (0, 0)
	SpriteCenter SpriteCoords
	// default is 0.0 & 0.0. This means the object is 0x0. Please make this 1.0
	StretchX float32
	StretchY float32
}

type drawingQueue struct {
	queue *list.List
	mu    sync.Mutex
}

var drawQueue *drawingQueue
var onceDrawQueue sync.Once

func initDrawQueue() {
	logger.LOG.Info().Msg("Creating new draw queue")
	drawQueue = new(drawingQueue)
	drawQueue.queue = list.New()
}

func GetDrawQueue() *drawingQueue {
	onceDrawQueue.Do(initDrawQueue)
	return drawQueue
}

// Should never be called concurrently because it *COULD* use glfw/gl
// which must be called from main thread
func CreateSprite(initParams *SpriteInitParams) (*Sprite, error) {
	logger.LOG.Info().Msg("Creating new sprite")

	sprite := Sprite{}
	sprite.lazyDeletionMark.Store(false)

	var err error
	sprite.shaderId, err = getShader(initParams.ShaderRelPaths)
	if err != nil {
		return nil, err
	}
	sprite.vao, err = getVAO(initParams.TextureCoords)
	if err != nil {
		return nil, err
	}
	sprite.Tex, err = getTexture(initParams.TextureRelPath, initParams.TextureCoords)
	if err != nil {
		return nil, err
	}

	sprite.Tex.DimX *= initParams.StretchX
	sprite.Tex.DimY *= initParams.StretchY
	// default position in screen: top left (0, 0)
	sprite.ScreenCenter = initParams.ScreenCenter
	// default origin of sprite: upper left (0, 0)
	sprite.SpriteCenter = initParams.SpriteCenter

	return &sprite, nil
}

func (dq *drawingQueue) AddToQueue(w weak.Pointer[Sprite]) {
	dq.mu.Lock()
	defer dq.mu.Unlock()

	// checking if this is already in the queue
	listElem := dq.queue.Front()
	for listElem != nil {
		nextListElem := listElem.Next()

		weakSprite, ok := listElem.Value.(weak.Pointer[Sprite])
		if !ok {
			logger.LOG.Fatal().Msg("Saw a non-drawQueueObject in the Draw Queue.")
		}

		if weakSprite.Value() == nil {
			logger.LOG.Debug().Msg("Removed a nil draw Object (object got Gc'd)")
			dq.queue.Remove(listElem)
		} else if weakSprite == w {
			logger.LOG.Warn().Msg("object already in queue. Not adding again.")
			dq.queue.Remove(listElem)
			return
		}

		listElem = nextListElem
	}

	dq.queue.PushFront(w)

	if dq.queue.Len() > 100 {
		logger.LOG.Warn().Msgf("Draw Queue is getting long. Len: %v", dq.queue.Len())
	}
	logger.LOG.Debug().Msgf(
		"Added draw object to the draw queue (ShaderID: %v, TextureID: %v): %v",
		w.Value().shaderId,
		w.Value().Tex.textureId,
		w,
	)
}

func (dq *drawingQueue) RemoveFromQueue(w weak.Pointer[Sprite]) (ok bool) {
	logger.LOG.Debug().Msgf(
		"Manually removing object from the drawQueue (ShaderID: %v, TextureID: %v): %v",
		w.Value().shaderId,
		w.Value().Tex.textureId,
		w,
	)

	dq.mu.Lock()
	defer dq.mu.Unlock()

	listElem := dq.queue.Front()
	for listElem != nil {
		nextListElem := listElem.Next()

		weakSprite, ok := listElem.Value.(weak.Pointer[Sprite])
		if !ok {
			logger.LOG.Fatal().Msg("Saw a non-drawQueueObject in the Draw Queue.")
			return ok
		}

		if weakSprite.Value() == nil {
			logger.LOG.Debug().Msg("Removed a nil draw Object (object got Gc'd)")
			dq.queue.Remove(listElem)
		} else if weakSprite == w {
			dq.queue.Remove(listElem)
			return true
		}

		listElem = nextListElem
	}
	logger.LOG.Debug().Msg("Couldn't find sprite to remove from Draw Queue")
	return true
}

// should always be called in the main thread (glfw & gl)
func (dq *drawingQueue) Draw() {
	listElem := dq.queue.Front()
	for listElem != nil {
		nextListElem := listElem.Next()

		weakSprite, ok := listElem.Value.(weak.Pointer[Sprite])
		if !ok {
			logger.LOG.Fatal().Msg("Saw a non-drawQueueObject in the Draw Queue.")
		}

		strongSprite := weakSprite.Value()
		if strongSprite == nil || strongSprite.IsNil() {
			logger.LOG.Debug().Msg("Removed a nil draw Object (object got Gc'd)")
			dq.queue.Remove(listElem)
		} else {
			// The graphics card thinks the sprite center is the bottom left.
			// The minus on the Y is because the Y coordinate direction is flipped in openGL,
			// meaning its a right-handed system. So the bottom left is 0, 0
			openGlScreenCenter := ScreenCoords{
				X: strongSprite.ScreenCenter.X - strongSprite.SpriteCenter.X*strongSprite.Tex.DimX,
				Y: strongSprite.ScreenCenter.Y - strongSprite.SpriteCenter.Y*strongSprite.Tex.DimY,
			}

			gl.UseProgram(strongSprite.shaderId)
			setTransform(strongSprite.shaderId, openGlScreenCenter)
			setScale(strongSprite.shaderId, strongSprite.Tex.DimX, strongSprite.Tex.DimY)

			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, strongSprite.Tex.textureId)

			gl.BindVertexArray(strongSprite.vao)
			gl.DrawArrays(gl.TRIANGLES, 0, 6)
			gl.BindVertexArray(0)
		}
		listElem = nextListElem
	}
}

func (s *Sprite) SpriteCoordsToScreenCoords(spriteCoords SpriteCoords) ScreenCoords {
	return ScreenCoords{
		X: s.ScreenCenter.X + (s.SpriteCenter.X-s.SpriteCenter.X)*s.Tex.DimX,
		Y: s.ScreenCenter.Y + (s.SpriteCenter.Y-s.SpriteCenter.Y)*s.Tex.DimY,
	}
}
