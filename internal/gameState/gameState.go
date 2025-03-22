package gameState

import (
	"sync"
	"sync/atomic"

	"github.com/PatrickKoch07/game-proj/internal/logger"
)

type gameState struct {
	currentState map[Flag]*atomic.Int32
	futureState  map[Flag]*atomic.Int32
}

var currentGameState *gameState
var once sync.Once

func initCurrentGameState() {
	logger.LOG.Debug().Msg("Creating new GameState.")
	currentGameState = new(gameState)
	currentGameState.currentState = make(map[Flag]*atomic.Int32)
	currentGameState.futureState = make(map[Flag]*atomic.Int32)
}

// This will hold the game state as a map of Flags (int32) to int32.
// not quite a sync.map because we want to force the switch between the current/dirty map
func GetCurrentGameState() *gameState {
	once.Do(initCurrentGameState)
	return currentGameState
}

// Should only be called from the main thread. Not in any race against reading/writing
func (gs *gameState) UpdateCurrentContext() {
	for key, item := range gs.futureState {
		gs.currentState[key] = item
	}
	gs.futureState = make(map[Flag]*atomic.Int32)
}

// Should never be in a race against writing (diff maps).
// So, should be thread safe for concurrent reading
func (gs *gameState) GetFlagValue(flag Flag) (int32, bool) {
	val, ok := gs.currentState[flag]
	if val == nil {
		return 0, false
	}
	return val.Load(), ok
}

func (gs *gameState) SetFlagValue(flag Flag, value int32) {
	_, ok := gs.futureState[flag]
	if !ok {
		gs.futureState[flag] = new(atomic.Int32)
	}
	gs.futureState[flag].Store(value)
}

// this will hold if the UI said to close, or if a scene change was requested
// this will also hold past character actions that could affect the future
// because of this, will be used to init scenes and objects anywhere
