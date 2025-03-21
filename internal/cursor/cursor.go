package cursor

// package level singleton because many components need to read from cursor
// Only want exactly one cursor per game (if any)

import (
	"weak"
	"sync"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
	"github.com/PatrickKoch07/game-proj/internal/utils"

	"github.com/go-gl/glfw/v3.3/glfw"
)

var gameCursor *sprites.Sprite
var once sync.Once

func GetCursor() *sprites.Sprite {
	once.Do(initCursor)
	return gameCursor
}

func GetCursorScreenPosition() (float32, float32) {
	return GetCursor().ScreenX, GetCursor().ScreenY
}

func UpdateMousePosCallback(w *glfw.Window, xpos float64, ypos float64) {
	// called on the main thread from GLFW poll events (don't worry about concurrency)

	if gameCursor == nil {
		return
	}
	// virtual mouse position in opengl
	w.SetCursorPos(0, 0)
	// logger.LOG.Debug().Msgf("Mouse moved (%v, %v)", xpos, ypos)

	// please don't resize the window solved by gameState TODO
	GetCursor().ScreenX = utils.Clamp(GetCursor().ScreenX+float32(xpos), 0.0, 1280.0)
	GetCursor().ScreenY = utils.Clamp(GetCursor().ScreenY+float32(ypos), 0.0, 960.0)
	// logger.LOG.Debug().Msgf("Mouse at (%v, %v)", GetCursor().Sprite.ScreenX, GetCursor().Sprite.ScreenY)
}

func initCursor() {
	logger.LOG.Info().Msg("Creating new cursor")

	// sprites.MakeShader(
	// 	sprites.ShaderFiles{
	// 		VertexPath:   "cursorShader.vs",
	// 		FragmentPath: "alphaTextureShader.fs",
	// 	})
	// sprites.MakeTexture("ui/cursor.png")
	// sprites.MakeVAO(sprites.TexCoordOneSpritePerImg)

	sprite, err := sprites.CreateSprite(
		&sprites.SpriteInitParams{
			ShaderRelPaths: sprites.ShaderFiles{
				VertexPath:   "cursorShader.vs",
				FragmentPath: "alphaTextureShader.fs",
			},
			TextureRelPath: "ui/cursor.png",
			TextureCoords:  sprites.TexCoordOneSpritePerImg,
			// inital callbacks make cursor jump to center of screen so this makes it not visual
			ScreenX: -100.0,
			ScreenY: -100.0,
			// ScreenX:       0.0,
			// ScreenY:       0.0,
			SpriteOriginX: 0.0,
			SpriteOriginY: 0.0,
			StretchX:      1.0,
			StretchY:      1.0,
		},
	)
	if err != nil {
		logger.LOG.Error().Msg("Cursor failed to be made.")
		logger.LOG.Error().Err(err).Msg("")
		return
	}
	gameCursor = sprite

	sprites.AddToDrawingQueue(weak.Make(gameCursor))
}
