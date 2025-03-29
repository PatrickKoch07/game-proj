package gameUi

import (
	"errors"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/audio"
	"github.com/PatrickKoch07/game-proj/internal/cursor"
	"github.com/PatrickKoch07/game-proj/internal/inputs"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

type button struct {
	Sprite        *sprites.Sprite
	AudioPlayer   audio.Player
	inputListener inputs.InputListener
	OnPress       func()
	OnRelease     func()
}

func CreateButton(height, width, screenX, screenY float32) (*button, error) {
	b := new(button)

	sprite, err := sprites.CreateSprite(
		&sprites.SpriteInitParams{
			ShaderRelPaths: sprites.ShaderFiles{
				VertexPath:   "uiShader.vs",
				FragmentPath: "alphaTextureShader.fs",
			},
			TextureRelPath: "ui/button.png",
			TextureCoords:  sprites.TexCoordOneSpritePerImg,
			ScreenCenter:   sprites.ScreenCoords{X: screenX, Y: screenY},
			SpriteCenter:   sprites.SpriteCoords{X: 0.0, Y: 0.0},
			// below two lines technically not needed since we manually change Tex Dim anyway
			StretchX: 1.0,
			StretchY: 1.0,
		},
	)
	if err != nil {
		return nil, err
	}
	b.Sprite = sprite
	b.Sprite.Tex.DimX = width
	b.Sprite.Tex.DimY = height

	b.OnPress = func() {}
	b.OnRelease = func() {}

	b.inputListener = inputs.InputListener(b)
	ok := inputs.GetInputManager().Subscribe(inputs.LMB, weak.Make(&b.inputListener))
	if !ok {
		return nil, errors.New("failed to subscribe")
	}
	b.AudioPlayer, err = audio.CreatePlayer("assets/audio/buttonPress.mp3")
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (b *button) UnsubInput() {
	inputs.GetInputManager().Unsubscribe(inputs.LMB, weak.Make(&b.inputListener))
}

func (b *button) OnKeyAction(keyAction inputs.KeyAction) {
	mScreenPos := cursor.GetCursor().ScreenCenter
	if mScreenPos.X <= b.Sprite.ScreenCenter.X {
		return
	}
	if mScreenPos.X >= b.Sprite.ScreenCenter.X+b.Sprite.Tex.DimX {
		return
	}

	if mScreenPos.Y <= b.Sprite.ScreenCenter.Y {
		return
	}
	if mScreenPos.Y >= b.Sprite.ScreenCenter.Y+b.Sprite.Tex.DimY {
		return
	}

	b.AudioPlayer.Play()

	if keyAction.Action == inputs.Press {
		logger.LOG.Debug().Msgf(
			"Mouse pressed at (%v, %v)",
			mScreenPos.X,
			mScreenPos.Y,
		)
		b.OnPress()
	}

	if keyAction.Action == inputs.Release {
		logger.LOG.Debug().Msgf(
			"Mouse released at (%v, %v)",
			mScreenPos.X,
			mScreenPos.Y,
		)
		b.OnRelease()
	}
}
