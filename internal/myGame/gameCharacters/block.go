package gameCharacters

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/audio"
	"github.com/PatrickKoch07/game-proj/internal/camera"
	"github.com/PatrickKoch07/game-proj/internal/characters"
	"github.com/PatrickKoch07/game-proj/internal/colliders"
	"github.com/PatrickKoch07/game-proj/internal/gameState"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/scenes"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

type Block struct {
	*characters.CollidableObject
}

func (b *Block) InitInstance() ([]scenes.GameObject, []*sprites.Sprite, []audio.Player, bool) {
	var Sprites []*sprites.Sprite = make([]*sprites.Sprite, 1)
	var AudioPlayers []audio.Player = make([]audio.Player, 0)
	var GameObjects []scenes.GameObject = make([]scenes.GameObject, 1)
	GameObjects[0] = b
	creationSuccess := true

	collider := colliders.Collider2D{
		Tags:             make([]gameState.Flag, 0),
		CenterCoords:     colliders.WorldCoords{X: 0.0, Y: 0.0},
		Width:            128.0,
		Height:           128.0,
		OnEnterCollision: func(c *colliders.Collider2D) { logger.LOG.Debug().Msg("block collided") },
		OnExitCollision:  func(c *colliders.Collider2D) { logger.LOG.Debug().Msg("block stopped colliding") },
		Block:            make([]gameState.Flag, 0),
		Ignore:           make([]gameState.Flag, 0),
		Parent:           &GameObjects[0],
	}
	// collider.Tags[0] = gameState.EnvironmentCollider
	colliderSprites := make([]*sprites.Sprite, 1)
	colliderSprite, err := sprites.CreateSprite(
		&sprites.SpriteInitParams{
			ShaderRelPaths: sprites.ShaderFiles{
				VertexPath:   "alphaTextureShader.vs",
				FragmentPath: "alphaTextureShader.fs",
			},
			TextureRelPath: "ui/button.png",
			TextureCoords:  sprites.TexCoordOneSpritePerImg,
			ScreenCenter:   camera.WorldCoordsToScreenCoords(collider.CenterCoords),
			SpriteCenter:   sprites.SpriteCoords{X: 0.5, Y: 0.5},
			StretchX:       1.0,
			StretchY:       1.0,
		},
	)
	colliderSprite.Tex.DimX = collider.Width
	colliderSprite.Tex.DimY = collider.Height
	colliderSprites[0] = colliderSprite
	b.CollidableObject = characters.CreateCollidableObject(&collider, colliderSprites)
	if err != nil {
		creationSuccess = false
	}

	Sprites[0] = b.Sprites[0]

	for _, sprite := range b.Sprites {
		if sprite == nil {
			logger.LOG.Error().Msg(
				"nil sprite encountered adding Block sprites to draw queue (init)",
			)
		} else {
			sprites.GetDrawQueue().AddToQueue(weak.Make(sprite))
		}
	}

	// GameObjects[0] = b
	return GameObjects, Sprites, AudioPlayers, creationSuccess
}

func (b *Block) Update() {
	return
}

func (b *Block) ShouldSkipUpdate() bool {
	return true
}

func (b *Block) Kill() {
	for _, sprite := range b.Sprites {
		sprites.GetDrawQueue().RemoveFromQueue(weak.Make(sprite))
	}
}

func (b *Block) IsDead() bool {
	return false
}
