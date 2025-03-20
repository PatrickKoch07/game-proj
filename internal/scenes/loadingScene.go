package scenes

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

func createLoadingScene() *Scene {
	loadingScene := new(Scene)
	loadingScene.Sprites = make([]*sprites.Sprite, 1)
	sprite, err := sprites.CreateSprite(
		&sprites.SpriteInitParams{
			ShaderRelPaths: sprites.ShaderFiles{
				VertexPath:   "uiShader.vs",
				FragmentPath: "alphaTextureShader.fs",
			},
			TextureRelPath: "ui/loadingScreen.png",
			TextureCoords:  sprites.TexCoordOneSpritePerImg,
			ScreenX:        0,
			ScreenY:        0,
			SpriteOriginX:  0,
			SpriteOriginY:  0,
			StretchX:       1,
			StretchY:       1,
		},
	)
	if err != nil {
		logger.LOG.Error().Err(err).Msg("")
	} else {
		loadingScene.Sprites[0] = sprite
	}
	loadingScene.Init = initLoadingScene

	return loadingScene
}

func initLoadingScene(loadingScene *Scene) {
	sprites.AddToDrawingQueue(weak.Make(loadingScene.Sprites[0]))
}
