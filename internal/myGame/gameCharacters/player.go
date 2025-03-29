package gameCharacters

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/audio"
	"github.com/PatrickKoch07/game-proj/internal/camera"
	"github.com/PatrickKoch07/game-proj/internal/characters"
	"github.com/PatrickKoch07/game-proj/internal/colliders"
	"github.com/PatrickKoch07/game-proj/internal/gameState"
	"github.com/PatrickKoch07/game-proj/internal/inputs"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/scenes"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

type Player struct {
	*characters.CollidableObject
	inputListener inputs.InputListener
	// temp
	death         bool
	baseVelocityX float32
	baseVelocityY float32
	movespeed     float32
}

// func createPlayer() *player {
// }

func (p *Player) InitInstance() ([]scenes.GameObject, []*sprites.Sprite, []audio.Player, bool) {
	var Sprites []*sprites.Sprite = make([]*sprites.Sprite, 0, 1)
	var AudioPlayers []audio.Player = make([]audio.Player, 0)
	var GameObjects []scenes.GameObject = make([]scenes.GameObject, 0, 1)
	GameObjects = append(GameObjects, p)

	p.death = false
	p.movespeed = 5.0
	creationSuccess := true

	collider := colliders.Collider2D{
		Tags:             make([]gameState.Flag, 0),
		CenterCoords:     colliders.WorldCoords{X: 300.0, Y: 300.0},
		Width:            32.0,
		Height:           32.0,
		OnEnterCollision: func(c *colliders.Collider2D) { logger.LOG.Debug().Msg("player collided") },
		OnExitCollision:  func(c *colliders.Collider2D) { logger.LOG.Debug().Msg("player stopped colliding") },
		Block:            make([]gameState.Flag, 1),
		Ignore:           make([]gameState.Flag, 0),
		Parent:           &GameObjects[0],
	}
	collider.Block[0] = gameState.EnvironmentCollider
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
	p.CollidableObject = characters.CreateCollidableObject(&collider, colliderSprites)
	if err != nil {
		creationSuccess = false
	}
	// sprite.Tex.DimX = 100
	// sprite.Tex.DimY = 100
	Sprites = append(Sprites, p.CollidableObject.Sprites...)

	p.inputListener = inputs.InputListener(p)
	ok := inputs.GetInputManager().Subscribe(inputs.KeyW, weak.Make(&p.inputListener))
	if !ok {
		creationSuccess = false
	}
	ok = inputs.GetInputManager().Subscribe(inputs.KeyA, weak.Make(&p.inputListener))
	if !ok {
		creationSuccess = false
	}
	ok = inputs.GetInputManager().Subscribe(inputs.KeyS, weak.Make(&p.inputListener))
	if !ok {
		creationSuccess = false
	}
	ok = inputs.GetInputManager().Subscribe(inputs.KeyD, weak.Make(&p.inputListener))
	if !ok {
		creationSuccess = false
	}

	for _, sprite := range Sprites {
		if sprite == nil {
			logger.LOG.Error().Msg(
				"nil sprite encountered adding Player sprites to draw queue (init)",
			)
		} else {
			sprites.GetDrawQueue().AddToQueue(weak.Make(sprite))
		}
	}

	// GameObjects = append(GameObjects, p)
	return GameObjects, Sprites, AudioPlayers, creationSuccess
}

func (p *Player) OnKeyAction(ka inputs.KeyAction) {
	if ka.Key == inputs.KeyW {
		if ka.Action != inputs.Release {
			p.baseVelocityY += p.movespeed
		} else {
			p.baseVelocityY -= p.movespeed
		}
	}
	if ka.Key == inputs.KeyA {
		if ka.Action != inputs.Release {
			p.baseVelocityX -= p.movespeed
		} else {
			p.baseVelocityX += p.movespeed
		}
	}
	if ka.Key == inputs.KeyS {
		if ka.Action != inputs.Release {
			p.baseVelocityY -= p.movespeed
		} else {
			p.baseVelocityY += p.movespeed
		}
	}
	if ka.Key == inputs.KeyD {
		if ka.Action != inputs.Release {
			p.baseVelocityX += p.movespeed
		} else {
			p.baseVelocityX -= p.movespeed
		}
	}
}

func (p *Player) Update() {
	p.MoveCharacter(
		colliders.WorldCoords{
			X: p.Collider.CenterCoords.X + p.baseVelocityX,
			Y: p.Collider.CenterCoords.Y + p.baseVelocityY,
		},
	)
}

func (p *Player) ShouldSkipUpdate() bool {
	return false
}

func (p *Player) Kill() {
	for _, sprite := range p.Sprites {
		sprites.GetDrawQueue().RemoveFromQueue(weak.Make(sprite))
	}
}

func (p *Player) IsDead() bool {
	return p.death
}
