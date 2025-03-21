package sprites

import (
	"container/list"
	"sync"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type Sprite struct {
	shaderId uint32
	Tex      texture
	vao      uint32
	// Where the origin of the object is on the screen (top left is 0.0, 0.0)
	ScreenX float32
	ScreenY float32
	// from 0.0 to 1.0, Where the origin of the object is on the sprite (top left is 0.0, 0.0)
	originSpriteX float32
	originSpriteY float32
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

type SpriteInitParams struct {
	// TextureCoords must be: Bottom left, top left, top right, bottom left, top right, bottom right
	ShaderRelPaths ShaderFiles
	TextureRelPath string
	TextureCoords  [12]float32
	ScreenX        float32
	ScreenY        float32
	SpriteOriginX  float32
	SpriteOriginY  float32
	StretchX       float32
	StretchY       float32
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
	// default position in screen: top left
	sprite.ScreenX = initParams.ScreenX
	sprite.ScreenY = initParams.ScreenY
	// origin of sprite: upper left
	sprite.originSpriteX = initParams.SpriteOriginX
	sprite.originSpriteY = initParams.SpriteOriginY

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
		if strongSprite == nil {
			logger.LOG.Debug().Msg("Removed a nil draw Object (object got Gc'd)")
			dq.queue.Remove(listElem)
		} else {
			screenX, screenY := strongSprite.getShaderOriginInScreenSpace()

			gl.UseProgram(strongSprite.shaderId)
			setTransform(strongSprite.shaderId, screenX, screenY)
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

// technically shouldn't be part of sprite struct? but oh well, nobody else should use this anyway
func (s *Sprite) getShaderOriginInScreenSpace() (x float32, y float32) {
	// shader origin is defined as bottom left.
	x = s.ScreenX - s.originSpriteX*s.Tex.DimX
	y = s.ScreenY - s.originSpriteY*s.Tex.DimY
	return x, y
}
