package gameObjects

import (
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

type GameObject interface {
	InitInstance() ([]*sprites.Sprite, bool)
	Update()
	ShouldSkipUpdate() bool
}
