package sprites

import (
	"container/list"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type DrawObject struct {
	// objects that wish to be draw should have this
	ShaderId  uint32
	TextureId uint32
	// other params to follow
}

type MoveObject interface {
	GetBottomLeftInScreenCoords() (float32, float32)
}

var drawQueue *list.List

type drawQueueObject struct {
	weakDrawObj weak.Pointer[DrawObject]
	weakMoveObj weak.Pointer[MoveObject]
}

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

func AddToDrawingQueue(wD weak.Pointer[DrawObject], wM weak.Pointer[MoveObject]) {
	drawQueueObj := drawQueueObject{weakDrawObj: wD, weakMoveObj: wM}

	drawQueue.PushFront(drawQueueObj)
	if drawQueue.Len() > 100 {
		logger.LOG.Warn().Msgf("Draw Queue is getting long. Len: %v", drawQueue.Len())
	}
	logger.LOG.Debug().Msgf(
		"Added draw object to the draw queue: %v. ShaderID: %v. TextureID: %v",
		wD,
		wD.Value().ShaderId,
		wD.Value().TextureId,
	)
}

func RemoveFromDrawingQueue(w weak.Pointer[DrawObject]) (ok bool) {
	logger.LOG.Debug().Msgf("Manually removing object from the drawQueue: %v", w)

	listElem := getDrawQueue().Front()
	for listElem != nil {
		nextListElem := listElem.Next()

		drawQueueObj, ok := listElem.Value.(drawQueueObject)
		if !ok {
			logger.LOG.Fatal().Msg("Saw a non-drawQueueObject in the Draw Queue.")
			return ok
		}

		drawObj := drawQueueObj.weakDrawObj
		if drawObj.Value() == nil {
			logger.LOG.Debug().Msg("Removed a nil draw Object (object got Gc'd)")
			getDrawQueue().Remove(listElem)
		} else if drawObj == w {
			getDrawQueue().Remove(listElem)
			return true
		}

		listElem = nextListElem
	}

	logger.LOG.Warn().Msgf("Object not found in the drawQueue: %v", w)
	return false
}

func DeleteShaders(shaderIds ...uint32) {
	for sid := range shaderIds {
		gl.DeleteProgram(uint32(sid))
	}
}

func DeleteTextures(textureIds ...uint32) {
	numTextures := int32(len(textureIds))
	gl.DeleteTextures(numTextures, &textureIds[0])
}

func DrawDrawQueue() {
	listElem := getDrawQueue().Front()
	for listElem != nil {
		nextListElem := listElem.Next()

		drawQueueObj, ok := listElem.Value.(drawQueueObject)
		if !ok {
			logger.LOG.Fatal().Msg("Saw a non-drawQueueObject in the Draw Queue.")
		}

		moveObj := drawQueueObj.weakMoveObj.Value()
		drawObj := drawQueueObj.weakDrawObj.Value()
		if drawObj == nil {
			logger.LOG.Debug().Msg("Removed a nil draw Object (object got Gc'd)")
			getDrawQueue().Remove(listElem)
		} else {
			if moveObj != nil {
				shaderId := drawObj.ShaderId
				screenX, screenY := (*moveObj).GetBottomLeftInScreenCoords()
				SetTransform(shaderId, screenX, screenY)
			}
			// use shader, use texture, use etc...
		}
		listElem = nextListElem
	}
}

func init() {
	initDrawQueue()
}

func initDrawQueue() {
	logger.LOG.Info().Msg("Creating draw queue")
	drawQueue = list.New()
}
