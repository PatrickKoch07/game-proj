package cursor

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
	"github.com/PatrickKoch07/game-proj/internal/utils"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type cursor struct {
	Sprite sprites.Sprite
}

var gameCursor *cursor

func GetCursor() *cursor {
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
	GetCursor().Sprite.ScreenX = utils.Clamp(GetCursor().Sprite.ScreenX+float32(xpos), 0.0, 1280.0)
	GetCursor().Sprite.ScreenY = utils.Clamp(GetCursor().Sprite.ScreenY+float32(ypos), 0.0, 960.0)
	// logger.LOG.Debug().Msgf("Mouse at (%v, %v)", GetCursor().Sprite.ScreenX, GetCursor().Sprite.ScreenY)
}

func InitCursor() {
	logger.LOG.Info().Msg("Creating new cursor")

	gameCursor = new(cursor)
	gameCursor.Sprite = sprites.Sprite{}

	shaderId, ok := sprites.MakeShader("cursorShader.vs", "alphaTextureShader.fs")
	if !ok {
		logger.LOG.Fatal().Msg("Shader for cursor failed to be made.")
	}
	textId, spriteWidth, spriteHeight, err := sprites.GenerateTexture("cursor.png")
	if err != nil {
		logger.LOG.Error().Err(err).Msg(". From cursor.")
	}

	gameCursor.Sprite.TextureId = textId
	gameCursor.Sprite.ShaderId = shaderId
	// default position in screen: top left
	gameCursor.Sprite.ScreenX = 0.0
	gameCursor.Sprite.ScreenY = 0.0
	// origin of sprite: upper left
	gameCursor.Sprite.OriginSpriteX = 0.0
	gameCursor.Sprite.OriginSpriteY = 0.0
	gameCursor.Sprite.SpriteHeight = float32(spriteHeight)
	gameCursor.Sprite.SpriteWidth = float32(spriteWidth)

	sprites.AddToDrawingQueue(weak.Make(&gameCursor.Sprite))
	// logger.LOG.Debug().Msgf(
	// 	"Created cursor with ShaderID: %v, TextureID: %v",
	// 	shaderId,
	// 	textId,
	// )
}
