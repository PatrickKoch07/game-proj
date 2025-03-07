package inputs

import (
	"container/list"
	"errors"
	"reflect"
	"weak"

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

type InputListener interface {
	OnKeyAction(glfw.Action)
}

type inputManager struct {
	keyStates      map[glfw.Key]KeyState
	keyActionQueue []KeyAction
	keyListeners   map[glfw.Key]*list.List
}

func GetInputManager() *inputManager {
	return singleton_im
}

func GetKeyState(key glfw.Key) (KeyState, bool) {
	value, ok := singleton_im.keyStates[key]
	return value, ok
}

func Subscribe(key glfw.Key, w weak.Pointer[InputListener]) bool {
	if _, ok := singleton_im.keyListeners[key]; !ok {
		logger.LOG.Error().Msgf("Trying to subscribe to bad key: %v", key)
		return ok
	}

	singleton_im.keyListeners[key].PushFront(w)
	if singleton_im.keyListeners[key].Len() > 10 {
		logger.LOG.Warn().Msgf("(Key: %v) Listener list has a lot of listeners (> 10).", key)
	}
	logger.LOG.Debug().Msgf("(Key: %v) Added subscriber: %v(%v)",
		key,
		w.Value(),
		reflect.TypeOf(*w.Value()),
	)
	return true
}

func Unsubscribe(key glfw.Key, w weak.Pointer[InputListener]) error {
	listElem := singleton_im.keyListeners[key].Front()
	for ; listElem != nil; listElem = listElem.Next() {
		// if we encounter nil valued elem, we delete. So should store next here.
		nextListElem := listElem.Next()

		switch listener := listElem.Value.(type) {
		case nil:
			logger.LOG.Debug().Msgf("(Key: %v) Removed nil listener", key)
			singleton_im.keyListeners[key].Remove(listElem)
		case weak.Pointer[InputListener]:
			strongListener := listener.Value()
			if strongListener == nil {
				logger.LOG.Debug().Msgf("(Key: %v) Removed nil listener", key)
				singleton_im.keyListeners[key].Remove(listElem)
			} else if listener == w {
				logger.LOG.Debug().Msgf("(Key: %v) Removing subscriber: %v(%v)",
					key,
					w.Value(),
					reflect.TypeOf(*w.Value()),
				)
				singleton_im.keyListeners[key].Remove(listElem)
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
	// for all Actions in input queue
	for ka, ok := k.dirtyPop(); ok; ka, ok = k.dirtyPop() {
		// notify all listeners of that key
		for listElem := k.keyListeners[ka.Key].Front(); listElem != nil; {
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
						"(Key: %v) Input Manager notifying: %v (%v)",
						ka.Key,
						listener,
						reflect.TypeOf(listener).Name(),
					)
					go (*strongListener).OnKeyAction(ka.Action)
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
	// because dirty pop
	k.keyActionQueue = make([]KeyAction, 0, inputManagerQueueSize)
}

func InputKeysCallback(
	w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	// Throw away repeat press case
	if action == glfw.Repeat {
		return
	}

	err := singleton_im.push(KeyAction{Key: key, Action: action})
	if err != nil {
		logger.LOG.Fatal().Err(err).Msg("Error in glfw to input queue.")
	}
}

// 10 seems like a large number for every frame's worth of inputs
const inputManagerQueueSize int = 10

var singleton_im *inputManager

func init() {
	singleton_im = new(inputManager)
	singleton_im.keyActionQueue = make([]KeyAction, 0, inputManagerQueueSize)
	singleton_im.keyStates = make(map[glfw.Key]KeyState)
	singleton_im.keyStates[glfw.KeyW] = Inactive
	singleton_im.keyStates[glfw.KeyA] = Inactive
	singleton_im.keyStates[glfw.KeyS] = Inactive
	singleton_im.keyStates[glfw.KeyD] = Inactive
	singleton_im.keyStates[glfw.KeyEscape] = Inactive
	singleton_im.keyListeners = make(map[glfw.Key]*list.List)
	singleton_im.keyListeners[glfw.KeyW] = list.New()
	singleton_im.keyListeners[glfw.KeyA] = list.New()
	singleton_im.keyListeners[glfw.KeyS] = list.New()
	singleton_im.keyListeners[glfw.KeyD] = list.New()
	singleton_im.keyListeners[glfw.KeyEscape] = list.New()
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
