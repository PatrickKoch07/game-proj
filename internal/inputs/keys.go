package inputs

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

type KeyState int

const (
	Inactive KeyState = 0
	Pressed  KeyState = 1
)

type Key glfw.Key

const (
	KeyW      Key = Key(glfw.KeyW)
	KeyA      Key = Key(glfw.KeyA)
	KeyS      Key = Key(glfw.KeyS)
	KeyD      Key = Key(glfw.KeyD)
	KeyEscape Key = Key(glfw.KeyEscape)
	LMB       Key = Key(glfw.MouseButton1*-1 - 2)
	RMB       Key = Key(glfw.MouseButton2*-1 - 2)
)

type Action glfw.Action

const (
	Press   Action = Action(glfw.Press)
	Release Action = Action(glfw.Release)
)

func MouseButtonToKey(m glfw.MouseButton) int {
	return (-1 * int(m)) - 2
}
