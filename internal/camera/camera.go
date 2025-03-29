package camera

import (
	"sync"

	"github.com/PatrickKoch07/game-proj/internal/colliders"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

type camera struct {
	// the center of the screen/camera in world coords
	WorldCenter  colliders.WorldCoords
	ScreenWidth  int
	ScreenHeight int
}

var currentCamera *camera
var once sync.Once

func GetCamera() *camera {
	if currentCamera == nil {
		once.Do(initCamera)
	}
	return currentCamera
}

func initCamera() {
	currentCamera = new(camera)
}

func InitializeCamera(screenWidth int, screenHeight int) {
	GetCamera().ScreenHeight = screenHeight
	GetCamera().ScreenWidth = screenWidth
	GetCamera().WorldCenter.X = 0.0
	GetCamera().WorldCenter.Y = 0.0
}

func ScreenCoordsToWorldCoords(screenCoords sprites.ScreenCoords) colliders.WorldCoords {
	// the plus is because the y direction is flipped for world coords (right handed system)
	return colliders.WorldCoords{
		X: GetCamera().WorldCenter.X - float32(GetCamera().ScreenWidth)/2.0 + screenCoords.X,
		Y: GetCamera().WorldCenter.Y + float32(GetCamera().ScreenHeight)/2.0 - screenCoords.Y,
	}
}

func WorldCoordsToScreenCoords(worldCoords colliders.WorldCoords) sprites.ScreenCoords {
	// the minus is because the y direction is flipped for world coords (right handed system)
	return sprites.ScreenCoords{
		X: (worldCoords.X - GetCamera().WorldCenter.X + float32(GetCamera().ScreenWidth)/2.0),
		Y: (-worldCoords.Y + GetCamera().WorldCenter.Y + float32(GetCamera().ScreenHeight)/2.0),
	}
}

// TODO some draw optimizations on where the camera is currently at, so we don't even bother to
// loop through drawing the sprites very much not in view

// TODO moving the camera methods & attaching it to some object
