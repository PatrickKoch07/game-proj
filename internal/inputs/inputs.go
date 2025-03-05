package inputs

import (
	"errors"
	"runtime"
	"reflect"

	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/PatrickKoch07/game-proj/internal/logger"
)

type KeyState int

const (
	Inactive KeyState = 0
	Pressed  KeyState = 1
)

type KeyAction struct {
	Key    glfw.Key
	Action glfw.Action
}

type inputManager struct {
	keyStates      map[glfw.Key]KeyState
	keyActionQueue []KeyAction
	listeners      map[glfw.Key][]func(glfw.Action)
}

func (k *inputManager) GetKeyState(key glfw.Key) (KeyState, bool) {
	value, ok := k.keyStates[key]
	return value, ok
}

func (k *inputManager) Subscribe(key glfw.Key, f func(glfw.Action)) bool {
	_, ok := k.listeners[key]
	if !ok {
		return ok
	}
	k.listeners[key] = append(k.listeners[key], f)
	return true
}

func (k *inputManager) Unsubscribe() {

}

func (k *inputManager) Notify() {
	for ka, ok := k.dirtyPop(); ok; ka, ok = k.dirtyPop() {
		for _, listenerFunc := range k.listeners[ka.Key] {
			logger.LOG.Debug().Msgf("Input Manager calling: %v",
				runtime.FuncForPC(reflect.ValueOf(listenerFunc).Pointer()).Name())
			go listenerFunc(ka.Action)
		}
	}
	// because dirty pop
	k.keyActionQueue = make([]KeyAction, 0, inputManagerQueueSize)
}

var INPUT_MANAGER *inputManager

func InputKeyCallback(
	w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	// Throw away repeat press case
	if action == glfw.Repeat {
		return
	}

	err := INPUT_MANAGER.push(KeyAction{Key: key, Action: action})
	if err != nil {
		logger.LOG.Fatal().Err(err).Msg("Error in glfw to input queue.")
	}
}

// 10 seems like a large number for every frame's worth of inputs
const inputManagerQueueSize int = 10

// 50 seems like an arbitrary HUGE number to me
const inputManagerListenerSize int = 50

func init() {
	INPUT_MANAGER = new(inputManager)
	INPUT_MANAGER.keyActionQueue = make([]KeyAction, 0, inputManagerQueueSize)
	INPUT_MANAGER.keyStates = make(map[glfw.Key]KeyState)
	INPUT_MANAGER.keyStates[glfw.KeyW] = Inactive
	INPUT_MANAGER.keyStates[glfw.KeyA] = Inactive
	INPUT_MANAGER.keyStates[glfw.KeyS] = Inactive
	INPUT_MANAGER.keyStates[glfw.KeyD] = Inactive
	INPUT_MANAGER.keyStates[glfw.KeyEscape] = Inactive
	INPUT_MANAGER.listeners = make(map[glfw.Key][]func(glfw.Action))
	INPUT_MANAGER.listeners[glfw.KeyW] = make([]func(glfw.Action), 0, inputManagerListenerSize)
	INPUT_MANAGER.listeners[glfw.KeyA] = make([]func(glfw.Action), 0, inputManagerListenerSize)
	INPUT_MANAGER.listeners[glfw.KeyS] = make([]func(glfw.Action), 0, inputManagerListenerSize)
	INPUT_MANAGER.listeners[glfw.KeyD] = make([]func(glfw.Action), 0, inputManagerListenerSize)
	INPUT_MANAGER.listeners[glfw.KeyEscape] = make([]func(glfw.Action), 0, inputManagerListenerSize)
}

func (k *inputManager) push(ka KeyAction) error {
	logger.LOG.Debug().Msgf("KeyPressQueue push() appended: %v", ka)
	k.keyActionQueue = append(k.keyActionQueue, ka)
	if len(k.keyActionQueue) == cap(k.keyActionQueue) {
		return errors.New("unexpectedly high number of inputs queued")
	}
	return nil
}

func (k *inputManager) dirtyPop() (ka KeyAction, ok bool) {
	/*
		The idea is that when someone else calls pop, they will do a game function depending on
		if the key was pressed or released. The concept of a key being held down should fall on
		them to define. => pressed == held, or is held something that activates after a small
		delay a la dark souls.
	*/
	// defer logger.LOG.Debug().Msgf("KeyPressQueue Pop() returns(%v): %v", ok, ka)
	// Loop until we can return a key state change.
	for {
		if len(k.keyActionQueue) == 0 {
			return KeyAction{}, false
		}

		ka = k.keyActionQueue[0]
		k.keyActionQueue = k.keyActionQueue[1:]

		switch ka.Action {
		case glfw.Release:
			k.keyStates[ka.Key] = Inactive
			return ka, true
		case glfw.Press:
			if k.keyStates[ka.Key] == Pressed {
				continue
			}
			k.keyStates[ka.Key] = Pressed
			return ka, true
		}
	}
}
