package sprites

import (
	"container/list"
	"unsafe"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type Sprite struct {
	// objects that wish to be draw should have this
	ShaderId  uint32
	TextureId uint32

	// other params to follow
	ScreenX float32
	ScreenY float32

	OriginSpriteX float32
	OriginSpriteY float32
	SpriteWidth   float32
	SpriteHeight  float32
}

var drawQueue *list.List

func getDrawQueue() *list.List {
	if drawQueue == nil {
		initDrawQueue()
	}
	return drawQueue
}

var rectangleSpritesVerts [24]float32 = [24]float32{
	// Position for the first two, Texture for the second two
	// Bottom left starting position
	0.0, 0.0, 0.0, 0.0,
	0.0, 1.0, 0.0, 1.0,
	1.0, 1.0, 1.0, 1.0,

	0.0, 0.0, 0.0, 0.0,
	1.0, 1.0, 1.0, 1.0,
	1.0, 0.0, 1.0, 0.0,
}

/*
	TEMP: I have the vbo & vao here assuming one png per sprite (no sprite sheets!!)
*/

var vao uint32

func InitRender() {
	logger.LOG.Info().Msg("Initializing sprite VAO & VBO")

	var vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)

	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	gl.BufferData(gl.ARRAY_BUFFER, 24*4, unsafe.Pointer(&rectangleSpritesVerts[0]), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, nil)
	gl.EnableVertexAttribArray(0)

	// unbind
	gl.BindVertexArray(0)
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
		"Added draw object to the draw queue. ShaderID: %v, TextureID: %v",
		w.Value().ShaderId,
		w.Value().TextureId,
	)
}

func RemoveFromDrawingQueue(w weak.Pointer[Sprite]) (ok bool) {
	logger.LOG.Debug().Msgf("Manually removing object from the drawQueue: %v", w)

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
			vertexArray := vao                              // strongSprite.VertexArray
			triangleCount := len(rectangleSpritesVerts) / 4 // strongSprite.TriangleCount
			screenX, screenY := strongSprite.getShaderOriginInScreenSpace()

			gl.UseProgram(strongSprite.ShaderId)
			SetTransform(strongSprite.ShaderId, screenX, screenY)
			SetScale(strongSprite.ShaderId, strongSprite.SpriteWidth, strongSprite.SpriteHeight)

			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, strongSprite.TextureId)

			gl.BindVertexArray(vertexArray)
			gl.DrawArrays(gl.TRIANGLES, 0, int32(triangleCount))
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

func (s *Sprite) getShaderOriginInScreenSpace() (x float32, y float32) {
	// shader origin is defined as bottom left.
	x = s.ScreenX - s.OriginSpriteX
	y = s.ScreenY + s.SpriteHeight - s.OriginSpriteY
	return x, y
}
