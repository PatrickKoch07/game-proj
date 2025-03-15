package cursor

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
	"github.com/PatrickKoch07/game-proj/internal/utils"

	"github.com/go-gl/glfw/v3.3/glfw"
)

var gameCursor *sprites.Sprite

func GetCursor() *sprites.Sprite {
	if gameCursor == nil {
		InitCursor()
	}
	return gameCursor
}

func UpdateMousePosCallback(w *glfw.Window, xpos float64, ypos float64) {
	// virtual mouse position in opengl
	w.SetCursorPos(0, 0)
	// logger.LOG.Debug().Msgf("Mouse moved (%v, %v)", xpos, ypos)

	// please don't resize the window
	GetCursor().ScreenX = utils.Clamp(GetCursor().ScreenX+float32(xpos), 0.0, 1280.0)
	GetCursor().ScreenY = utils.Clamp(GetCursor().ScreenY+float32(ypos), 0.0, 960.0)
	// logger.LOG.Debug().Msgf("Mouse at (%v, %v)", GetCursor().Sprite.ScreenX, GetCursor().Sprite.ScreenY)
}

func InitCursor() {
	logger.LOG.Info().Msg("Creating new cursor")

	sprite, err := sprites.CreateSprite(
		"cursorShader.vs",
		"alphaTextureShader.fs",
		"ui/cursor.png",
		sprites.TexCoordOneSpritePerImg,
		0.0,
		0.0,
		0.0,
		0.0,
	)
	if err != nil {
		logger.LOG.Error().Msg("Cursor failed to be made.")
		logger.LOG.Error().Err(err)
		return
	}
	gameCursor = sprite

	sprites.AddToDrawingQueue(weak.Make(gameCursor))
}
