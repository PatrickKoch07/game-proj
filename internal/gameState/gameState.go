package gameState

import (
	"sync"

	"github.com/PatrickKoch07/game-proj/internal/logger"
)

// not quite a sync.map because we want to force the switch between the current/dirty map
type gameState struct {
	currentState map[Flag]int
	futureState  map[Flag]int
}

var currentGameState *gameState
var once sync.Once

func initCurrentGameState() {
	logger.LOG.Debug().Msg("Creating new GameState.")
	currentGameState = new(gameState)
	currentGameState.currentState = make(map[Flag]int)
	currentGameState.futureState = make(map[Flag]int)
}

func GetCurrentGameState() *gameState {
	once.Do(initCurrentGameState)
	return currentGameState
}

func (gs *gameState) UpdateCurrentContext() {
	// run in the main loop
	for key, item := range gs.futureState {
		gs.currentState[key] = item
	}
	gs.futureState = make(map[Flag]int)
}

func (gs *gameState) GetFlagValue(flag Flag) (int, bool) {
	// needs a lock
	val, ok := gs.currentState[flag]
	return val, ok
}

func (gs *gameState) SetFlagValue(flag Flag, value int) {
	// needs a lock
	gs.futureState[flag] = value
}

// this will hold if the UI said to close, or if a scene change was requested
// this will also hold past character actions that could affect the future
// because of this, will be used to init scenes and objects anywhere
