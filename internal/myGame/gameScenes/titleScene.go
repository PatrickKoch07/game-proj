package gameScenes

import (
	"github.com/PatrickKoch07/game-proj/internal/myGame/gameUi"
	"github.com/PatrickKoch07/game-proj/internal/scenes"
)

func createTitleScene() *scenes.Scene {
	titleScene := new(scenes.Scene)
	mm := gameUi.MainMenu{}
	scenes.InitOnScene(titleScene, scenes.GameObject(mm), true)
	return titleScene
}
