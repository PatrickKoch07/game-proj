package characters

import (
	"github.com/PatrickKoch07/game-proj/internal/camera"
	"github.com/PatrickKoch07/game-proj/internal/colliders"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

// Holds sprites whose position are bound by some collider.
type CollidableObject struct {
	Collider *colliders.Collider2D
	Sprites  []*sprites.Sprite
}

func (c *CollidableObject) MoveCharacter(finalPoint colliders.WorldCoords) {
	finalCenter := c.Collider.MoveCollider(finalPoint)
	for _, sprite := range c.Sprites {
		sprite.ScreenCenter = camera.WorldCoordsToScreenCoords(finalCenter)
	}
}

func CreateCollidableObject(
	collider *colliders.Collider2D,
	sprites []*sprites.Sprite,
) *CollidableObject {
	c := new(CollidableObject)

	c.Collider = collider
	colliders.AddColliderToMaps(collider)
	c.Sprites = sprites

	return c
}
