package sprites

import (
	"container/list"
	"errors"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type Sprite struct {
	ShaderId uint32
	Tex      Texture
	VAO      uint32
	// Where the origin of the object is on the screen (top left is 0.0, 0.0)
	ScreenX float32
	ScreenY float32
	// from 0.0 to 1.0, Where the origin of the object is on the sprite (top left is 0.0, 0.0)
	OriginSpriteX float32
	OriginSpriteY float32
}

type SpriteInitParams struct {
	VertexShaderRelPath string
	FragShaderRelPath   string
	TextureRelPath      string
	TextureCoords       [12]float32
	ScreenX             float32
	ScreenY             float32
	SpriteOriginX       float32
	SpriteOriginY       float32
}

func CreateSprite(sP *SpriteInitParams) (*Sprite, error) {
	logger.LOG.Info().Msg("Creating new sprite")

	sprite := Sprite{}

	var ok bool
	sprite.ShaderId, ok = MakeShader(sP.VertexShaderRelPath, sP.FragShaderRelPath)
	if !ok {
		return nil, errors.New("error making shader")
	}
	var err error
	sprite.VAO = setVAO(sP.TextureCoords)
	sprite.Tex, err = GenerateTexture(sP.TextureRelPath)
	if err != nil {
		return nil, errors.New("error generating texture")
	}

	// default position in screen: top left
	sprite.ScreenX = sP.ScreenX
	sprite.ScreenY = sP.ScreenY
	// origin of sprite: upper left
	sprite.OriginSpriteX = sP.SpriteOriginX
	sprite.OriginSpriteY = sP.SpriteOriginY

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
		w.Value().ShaderId,
		w.Value().Tex.TextureId,
		w,
	)
}

func RemoveFromDrawingQueue(w weak.Pointer[Sprite]) (ok bool) {
	logger.LOG.Debug().Msgf(
		"Manually removing object from the drawQueue (ShaderID: %v, TextureID: %v): %v",
		w.Value().ShaderId,
		w.Value().Tex.TextureId,
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

	logger.LOG.Warn().Msgf("Object not found in the drawQueue: %v", w)
	return false
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

			gl.UseProgram(strongSprite.ShaderId)
			setTransform(strongSprite.ShaderId, screenX, screenY)
			setScale(strongSprite.ShaderId, strongSprite.Tex.DimX, strongSprite.Tex.DimY)

			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, strongSprite.Tex.TextureId)

			gl.BindVertexArray(strongSprite.VAO)
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
	x = s.ScreenX - s.OriginSpriteX*s.Tex.DimX
	y = s.ScreenY - s.OriginSpriteY*s.Tex.DimY
	return x, y
}
