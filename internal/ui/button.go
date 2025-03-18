package ui

import (
	"errors"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/cursor"
	"github.com/PatrickKoch07/game-proj/internal/inputs"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type button struct {
	Sprite        *sprites.Sprite
	inputListener inputs.InputListener
	OnPress       func(float32, float32)
	OnRelease     func(float32, float32)
}

func createButton(height, width, screenX, screenY float32) (*button, error) {
	logger.LOG.Info().Msg("Creating new button")

	b := new(button)

	//temp
	sprites.MakeShader(
		sprites.ShaderFiles{
			VertexPath:   "uiShader.vs",
			FragmentPath: "alphaTextureShader.fs",
		})
	sprites.MakeTexture("ui/button.png")
	sprites.MakeVAO(sprites.TexCoordOneSpritePerImg)
	//end temp

	sprite, err := sprites.CreateSprite(
		&sprites.SpriteInitParams{
			ShaderRelPaths: sprites.ShaderFiles{
				VertexPath:   "uiShader.vs",
				FragmentPath: "alphaTextureShader.fs",
			},
			TextureRelPath: "ui/button.png",
			TextureCoords:  sprites.TexCoordOneSpritePerImg,
			ScreenX:        screenX,
			ScreenY:        screenY,
			SpriteOriginX:  0.0,
			SpriteOriginY:  0.0,
		},
	)
	if err != nil {
		return nil, err
	}
	b.Sprite = sprite
	b.Sprite.Tex.DimX = width
	b.Sprite.Tex.DimY = height

	b.OnPress = func(_ float32, _ float32) {}
	b.OnRelease = func(_ float32, _ float32) {}

	b.inputListener = inputs.InputListener(b)
	ok := inputs.Subscribe(
		glfw.Key(inputs.MouseButtonToKey(glfw.MouseButton1)),
		weak.Make(&b.inputListener),
	)
	if !ok {
		return nil, errors.New("failed to subscribe")
	}
	return b, nil
}

func (b *button) OnKeyAction(action glfw.Action) {
	mX := cursor.GetCursor().ScreenX
	if mX <= b.Sprite.ScreenX {
		return
	}
	if mX >= b.Sprite.ScreenX+b.Sprite.Tex.DimX {
		return
	}

	mY := cursor.GetCursor().ScreenY
	if mY <= b.Sprite.ScreenY {
		return
	}
	if mY >= b.Sprite.ScreenY+b.Sprite.Tex.DimY {
		return
	}

	if action == glfw.Press {
		logger.LOG.Debug().Msgf(
			"Mouse pressed at (%v, %v)",
			mX,
			mY,
		)
		b.OnPress(cursor.GetCursor().ScreenX, cursor.GetCursor().ScreenY)
	}

	if action == glfw.Release {
		logger.LOG.Debug().Msgf(
			"Mouse released at (%v, %v)",
			mX,
			mY,
		)
		b.OnRelease(cursor.GetCursor().ScreenX, cursor.GetCursor().ScreenY)
	}
}
