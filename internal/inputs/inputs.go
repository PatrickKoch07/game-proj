package inputs

// Package level state held by private singleton initiated at program start.
// The main thread will use a function to call notify every frame.
// During a frame, on any thread, objects can subscribe or unsubscribe. (todo)

import (
	"container/list"
	"errors"
	"reflect"
	"sync"
	"weak"

	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/PatrickKoch07/game-proj/internal/logger"
)

type KeyAction struct {
	Key    Key
	Action Action
}

type InputListener interface {
	OnKeyAction(Action)
}

type inputManager struct {
	keyStates      map[Key]KeyState
	keyActionQueue []KeyAction
	keyListeners   map[Key]*list.List
}

// 10 seems like a large number for every frame's worth of inputs
const inputManagerQueueSize int = 10

var inputManagerObj *inputManager

func init() {
	initInputManager()
}

func initInputManager() {
	logger.LOG.Info().Msg("Creating new Input Manager!")

	inputManagerObj = new(inputManager)
	inputManagerObj.keyActionQueue = make([]KeyAction, 0, inputManagerQueueSize)

	inputManagerObj.keyStates = make(map[Key]KeyState)
	inputManagerObj.keyStates[KeyW] = Inactive
	inputManagerObj.keyStates[KeyA] = Inactive
	inputManagerObj.keyStates[KeyS] = Inactive
	inputManagerObj.keyStates[KeyD] = Inactive
	inputManagerObj.keyStates[KeyEscape] = Inactive
	inputManagerObj.keyStates[LMB] = Inactive
	inputManagerObj.keyStates[RMB] = Inactive

	inputManagerObj.keyListeners = make(map[Key]*list.List)
	inputManagerObj.keyListeners[KeyW] = list.New()
	inputManagerObj.keyListeners[KeyA] = list.New()
	inputManagerObj.keyListeners[KeyS] = list.New()
	inputManagerObj.keyListeners[KeyD] = list.New()
	inputManagerObj.keyListeners[KeyEscape] = list.New()
	inputManagerObj.keyListeners[LMB] = list.New()
	inputManagerObj.keyListeners[RMB] = list.New()
}

func GetInputManager() *inputManager {
	if inputManagerObj == nil {
		initInputManager()
	}
	return inputManagerObj
}

func (k *inputManager) GetKeyState(key Key) (KeyState, bool) {
	value, ok := k.keyStates[key]
	return value, ok
}

func (k *inputManager) Subscribe(key Key, w weak.Pointer[InputListener]) bool {
	if _, ok := k.keyListeners[key]; !ok {
		logger.LOG.Error().Msgf("Trying to subscribe to bad key: %v", key)
		return ok
	}

	k.keyListeners[key].PushFront(w)
	if k.keyListeners[key].Len() > 10 {
		logger.LOG.Warn().Msgf("(Key: %v) Listener list has a lot of listeners (> 10).", key)
	}
	logger.LOG.Debug().Msgf("(Key: %v) Added subscriber: %v(%v)",
		key,
		w.Value(),
		reflect.TypeOf(*w.Value()),
	)
	return true
}

func (k *inputManager) Unsubscribe(key Key, w weak.Pointer[InputListener]) error {
	listenerList, ok := k.keyListeners[key]
	if !ok {
		return errors.New("key does not exist")
	}

	for listElem := listenerList.Front(); listElem != nil; {
		// if we encounter nil valued elem, we delete. So should store next here.
		nextListElem := listElem.Next()

		switch listener := listElem.Value.(type) {
		case nil:
			logger.LOG.Debug().Msgf("(Key: %v) Removed nil listener", key)
			k.keyListeners[key].Remove(listElem)
		case weak.Pointer[InputListener]:
			strongListener := listener.Value()
			if strongListener == nil {
				logger.LOG.Debug().Msgf("(Key: %v) Removed nil listener", key)
				k.keyListeners[key].Remove(listElem)
			} else if listener == w {
				logger.LOG.Debug().Msgf("(Key: %v) Removing subscriber: %v(%v)",
					key,
					w.Value(),
					reflect.TypeOf(*w.Value()),
				)
				k.keyListeners[key].Remove(listElem)
				return nil
			}
		default:
			logger.LOG.Fatal().Msgf("(Key: %v) Found listener not a weakptr to InputListener. %v",
				key,
				listener,
			)
		}

		listElem = nextListElem
	}

	logger.LOG.Warn().Msgf("(Key: %v) Failed to remove listener. Not found: %v(%v)",
		key,
		w.Value(),
		reflect.TypeOf(*w.Value()),
	)
	return errors.New("no listener to be removed")
}

func (k *inputManager) Notify() {
	var wg sync.WaitGroup
	// for all Actions in input queue
	for ka, ok := k.dirtyPop(); ok; ka, ok = k.dirtyPop() {
		listenerQueue, ok := k.keyListeners[ka.Key]
		if !ok {
			continue
		}

		// notify all listeners of that key
		for listElem := listenerQueue.Front(); listElem != nil; {
			// if we encounter nil valued elem, we delete. So should store next here.
			nextListElem := listElem.Next()

			switch listener := listElem.Value.(type) {
			case nil:
				logger.LOG.Debug().Msgf("(Key: %v) Removed nil listener", ka.Key)
				k.keyListeners[ka.Key].Remove(listElem)
			case weak.Pointer[InputListener]:
				strongListener := listener.Value()
				if strongListener == nil {
					logger.LOG.Debug().Msgf("(Key: %v) Removed nil listener", ka.Key)
					k.keyListeners[ka.Key].Remove(listElem)
				} else {
					logger.LOG.Debug().Msgf(
						"(Key: %v) Input Manager notifying: %v",
						ka.Key,
						strongListener,
					)

					wg.Add(1)
					go func() { defer wg.Done(); (*strongListener).OnKeyAction(ka.Action) }()
				}
			default:
				logger.LOG.Fatal().Msgf(
					"(Key: %v) Found listener not a weakptr to InputListener: %v (%v)",
					ka.Key,
					listener,
					reflect.TypeOf(listener).Name(),
				)
			}

			listElem = nextListElem
		}
	}
	wg.Wait()

	// because dirty pop
	k.keyActionQueue = make([]KeyAction, 0, inputManagerQueueSize)
}

func InputKeysCallback(
	w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	// Throw away repeat press case
	if action == glfw.Repeat {
		return
	}

	err := GetInputManager().push(KeyAction{Key: Key(key), Action: Action(action)})
	if err != nil {
		logger.LOG.Fatal().Err(err).Msg("Error in glfw to input queue.")
	}
}

func InputMouseCallback(
	w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	// Throw away repeat press case
	if action == glfw.Repeat {
		return
	}

	err := GetInputManager().push(KeyAction{Key: Key(MouseButtonToKey(button)), Action: Action(action)})
	if err != nil {
		logger.LOG.Fatal().Err(err).Msg("Error in glfw to input queue.")
	}
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
		case Release:
			k.keyStates[ka.Key] = Inactive
			return ka, true
		case Press:
			if k.keyStates[ka.Key] == Pressed {
				continue
			}
			k.keyStates[ka.Key] = Pressed
			return ka, true
		}
	}
}
