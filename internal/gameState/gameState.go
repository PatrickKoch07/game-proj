package gameState

import (
	"sync"

	"github.com/PatrickKoch07/game-proj/internal/logger"
)

type gameState struct {
	currentState map[string]uint8
	futureState  map[string]uint8
}

var currentGameState *gameState
var once sync.Once

func initCurrentGameState() {
	logger.LOG.Debug().Msg("Creating new GameState.")
	currentGameState = new(gameState)
	currentGameState.currentState = make(map[string]uint8)
	currentGameState.futureState = make(map[string]uint8)
}

func GetCurrentGameState() *gameState {
	once.Do(initCurrentGameState)
	return currentGameState
}

func UpdateCurrentContext() {

}

func GetFlagValue(flag string) (uint8, bool) {
	val, ok := currentGameState.currentState[flag]
	return val, ok
}

func SetFlagValue(flag string, value uint8) {
	currentGameState.futureState[flag] = value
}

// this will hold if the UI said to close, or if a scene change was requested
// this will also hold past character actions that could affect the future
// because of this, will be used to init scenes and objects anywhere
