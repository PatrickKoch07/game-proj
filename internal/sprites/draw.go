package sprites

import (
	"container/list"
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

var drawQueue *list.List

func getDrawQueue() *list.List {
	if drawQueue == nil {
		initDrawQueue()
	}
	return drawQueue
}

/*
BIG NOTE ABOUT THE FOLLOWING TWO FUNCTIONS (ADD TO DRAWING QUEUE & REMOVE FROM DRAWING QUEUE):

	The idea is that these two functions would be used whenever the player enters the start and
	end area of a zone. Ideally, they would hit a loading area for the next zone. Then they would
	hit a deload zone from the previous region. Some game state would keep track of what 'zone'
	they are in and what operation to do when they cross any boundary.
*/
func AddToDrawingQueue(w weak.Pointer[Sprite]) {
	drawQueue.PushFront(w)
	if drawQueue.Len() > 100 {
		logger.LOG.Warn().Msgf("Draw Queue is getting long. Len: %v", drawQueue.Len())
	}
	logger.LOG.Debug().Msgf(
		"Added draw object to the draw queue (ShaderID: %v, TextureID: %v): %v",
		w.Value().shaderId,
		w.Value().Tex.textureId,
		w,
	)
}

func RemoveFromDrawingQueue(w weak.Pointer[Sprite]) (ok bool) {
	logger.LOG.Debug().Msgf(
		"Manually removing object from the drawQueue (ShaderID: %v, TextureID: %v): %v",
		w.Value().shaderId,
		w.Value().Tex.textureId,
		w,
	)

	listElem := getDrawQueue().Front()
	for listElem != nil {
		nextListElem := listElem.Next()

		weakSprite, ok := listElem.Value.(weak.Pointer[Sprite])
		if !ok {
			logger.LOG.Fatal().Msg("Saw a non-drawQueueObject in the Draw Queue.")
			return ok
		}

		if weakSprite.Value() == nil {
			logger.LOG.Debug().Msg("Removed a nil draw Object (object got Gc'd)")
			getDrawQueue().Remove(listElem)
		} else if weakSprite == w {
			getDrawQueue().Remove(listElem)
			return true
		}

		listElem = nextListElem
	}
	logger.LOG.Debug().Msg("Couldn't find sprite to remove from Draw Queue")
	return true
}

func DrawDrawQueue() {
	listElem := getDrawQueue().Front()
	for listElem != nil {
		nextListElem := listElem.Next()

		weakSprite, ok := listElem.Value.(weak.Pointer[Sprite])
		if !ok {
			logger.LOG.Fatal().Msg("Saw a non-drawQueueObject in the Draw Queue.")
		}

		strongSprite := weakSprite.Value()
		if strongSprite == nil {
			logger.LOG.Debug().Msg("Removed a nil draw Object (object got Gc'd)")
			getDrawQueue().Remove(listElem)
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

func init() {
	initDrawQueue()
}

func initDrawQueue() {
	logger.LOG.Info().Msg("Creating new draw queue")
	drawQueue = list.New()
}

// technically shouldn't be part of sprite struct? but oh well, nobody else should use this anyway
func (s *Sprite) getShaderOriginInScreenSpace() (x float32, y float32) {
	// shader origin is defined as bottom left.
	x = s.ScreenX - s.originSpriteX*s.Tex.DimX
	y = s.ScreenY - s.originSpriteY*s.Tex.DimY
	return x, y
}
