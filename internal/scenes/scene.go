package scenes

import (
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

type scene struct {
	spriteObjs []sprites.Sprite
}

var TitleScreen scene
var WorldScreen scene

func init() {
}

func (s *scene) InitScene() {

}