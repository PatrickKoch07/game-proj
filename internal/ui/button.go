package ui

import (
	"errors"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/cursor"
	"github.com/PatrickKoch07/game-proj/internal/inputs"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

type button struct {
	Sprite        *sprites.Sprite
	inputListener inputs.InputListener
	OnPress       func()
	OnRelease     func()
}

func CreateButton(height, width, screenX, screenY float32) (*button, *sprites.Sprite, error) {
	b := new(button)

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
		return nil, sprite, err
	}
	b.Sprite = sprite
	b.Sprite.Tex.DimX = width
	b.Sprite.Tex.DimY = height

	b.OnPress = func() {}
	b.OnRelease = func() {}

	b.inputListener = inputs.InputListener(b)
	ok := inputs.GetInputManager().Subscribe(inputs.LMB, weak.Make(&b.inputListener))
	if !ok {
		return nil, sprite, errors.New("failed to subscribe")
	}
	return b, sprite, nil
}

func (b *button) OnKeyAction(action inputs.Action) {
	mX, mY := cursor.GetCursorScreenPosition()
	if mX <= b.Sprite.ScreenX {
		return
	}
	if mX >= b.Sprite.ScreenX+b.Sprite.Tex.DimX {
		return
	}

	if mY <= b.Sprite.ScreenY {
		return
	}
	if mY >= b.Sprite.ScreenY+b.Sprite.Tex.DimY {
		return
	}

	if action == inputs.Press {
		logger.LOG.Debug().Msgf(
			"Mouse pressed at (%v, %v)",
			mX,
			mY,
		)
		b.OnPress()
	}

	if action == inputs.Release {
		logger.LOG.Debug().Msgf(
			"Mouse released at (%v, %v)",
			mX,
			mY,
		)
		b.OnRelease()
	}
}
