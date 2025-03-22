package gameScenes

import (
	"github.com/PatrickKoch07/game-proj/internal/gameState"
	"github.com/PatrickKoch07/game-proj/internal/scenes"
)

func GetSceneMap() map[gameState.Flag]func() *scenes.Scene {
	sceneMap := make(map[gameState.Flag]func() *scenes.Scene)
	sceneMap[gameState.LoadingScene] = createLoadingScene
	sceneMap[gameState.TitleScene] = createTitleScene
	sceneMap[gameState.WorldScene] = createWorldScene
	return sceneMap
}
