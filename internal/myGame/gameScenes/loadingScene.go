package gameScenes

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/scenes"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

func createLoadingScene() *scenes.Scene {
	logger.LOG.Info().Msg("Making loadingScene")

	loadingScene := new(scenes.Scene)
	loadingScene.Sprites = make([]*sprites.Sprite, 1)
	sprite, err := sprites.CreateSprite(
		&sprites.SpriteInitParams{
			ShaderRelPaths: sprites.ShaderFiles{
				VertexPath:   "uiShader.vs",
				FragmentPath: "alphaTextureShader.fs",
			},
			TextureRelPath: "ui/loadingScreen.png",
			TextureCoords:  sprites.TexCoordOneSpritePerImg,
			ScreenCenter:   sprites.ScreenCoords{X: 0.0, Y: 0.0},
			SpriteCenter:   sprites.SpriteCoords{X: 0.0, Y: 0.0},
			StretchX:       1,
			StretchY:       1,
		},
	)
	if err != nil {
		logger.LOG.Error().Err(err).Msg("")
	} else {
		loadingScene.Sprites[0] = sprite
	}
	sprites.GetDrawQueue().AddToQueue(weak.Make(loadingScene.Sprites[0]))

	return loadingScene
}
