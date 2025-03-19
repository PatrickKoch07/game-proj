package scenes

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
	"github.com/PatrickKoch07/game-proj/internal/ui"
)

var titleScene *Scene

func GetTitleScene() *Scene {
	if titleScene == nil {
		createTitleScene()
	}
	return titleScene
}

func createTitleScene() {
	titleScene = new(Scene)
	titleScene.Init = initTitleScene
}

func initTitleScene() {
	// spriteInits := make([]sprites.SpriteInitParams, 2)
	// spriteInits[0] = sprites.SpriteInitParams{
	// 	ShaderRelPaths: sprites.ShaderFiles{
	// 		VertexPath:   "uiShader.vs",
	// 		FragmentPath: "alphaTextureShader.fs",
	// 	},
	// 	TextureRelPath: "ui/button.png",
	// 	TextureCoords:  sprites.TexCoordOneSpritePerImg,
	// 	ScreenX:        512,
	// 	ScreenY:        640,
	// 	SpriteOriginX:  0.0,
	// 	SpriteOriginY:  0.0,
	// 	StretchX:       10.0,
	// 	StretchY:       2.0,
	// }
	// spriteInits[1] = sprites.SpriteInitParams{
	// 	ShaderRelPaths: sprites.ShaderFiles{
	// 		VertexPath:   "uiShader.vs",
	// 		FragmentPath: "alphaTextureShader.fs",
	// 	},
	// 	TextureRelPath: "ui/button.png",
	// 	TextureCoords:  sprites.TexCoordOneSpritePerImg,
	// 	ScreenX:        512,
	// 	ScreenY:        756,
	// 	SpriteOriginX:  0.0,
	// 	SpriteOriginY:  0.0,
	// 	StretchX:       10.0,
	// 	StretchY:       2.0,
	// }
	mm := ui.MainMenu{}
	// I really don't like this,
	// should be a better solution than reallllly remembering to do this before the init
	mm.PlayButtonFunc = switchScene
	buttonSprites, ok := mm.InitInstance()
	if !ok {
		logger.LOG.Error().Msg("Issue initializing main menu")
	} else {
		for _, sprite := range buttonSprites {
			sprites.AddToDrawingQueue(weak.Make(sprite))
		}
		GetTitleScene().gameObjects = append(GetTitleScene().gameObjects, mm)
		GetTitleScene().sprites = append(GetTitleScene().sprites, buttonSprites...)
	}
}

func switchScene() {
	logger.LOG.Debug().Msg("Starting switch scene process")
	SetNextScene(GetWorldScene())
}
