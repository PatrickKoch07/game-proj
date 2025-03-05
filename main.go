package main

import (
	"runtime"

	"github.com/PatrickKoch07/game-proj/internal/inputs"
	"github.com/PatrickKoch07/game-proj/internal/logger"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	// for glfw
	runtime.LockOSThread()
}

func main() {
	logger.LOG.Info().Msg("Hello World")

	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(640, 480, "Testing", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	window.SetFramebufferSizeCallback(framebufferResizeCallback)
	window.SetKeyCallback(inputs.InputKeyCallback)

	inputs.INPUT_MANAGER.Subscribe(glfw.KeyW, printToWorld)

	for !window.ShouldClose() {
		// rendering
		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// swaping and polling
		window.SwapBuffers()
		glfw.PollEvents()

		inputs.INPUT_MANAGER.Notify()
	}
}

func framebufferResizeCallback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func printToWorld(a glfw.Action) {
	if a == glfw.Press {
		logger.LOG.Debug().Msg("Hello World")
	}
	if a == glfw.Release {
		logger.LOG.Debug().Msg("Goodby World")
	}
}
