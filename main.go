package main

import (
	"runtime"
	"time"
	"weak"

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
	window.SetKeyCallback(inputs.InputKeysCallback)

	myDummyObj := new(dummyObj)
	myDummyObj.s = "keep me"
	myDummyObj.i = inputs.InputListener(myDummyObj)
	inputs.Subscribe(glfw.KeyW, weak.Make(&myDummyObj.i))

	mySecondObj := new(dummyObj)
	mySecondObj.s = "throw me away"
	mySecondObj.i = inputs.InputListener(mySecondObj)
	inputs.Subscribe(glfw.KeyW, weak.Make(&mySecondObj.i))

	go unsubLater(myDummyObj)
	go gcLater()

	for !window.ShouldClose() {
		// rendering
		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// swaping and polling
		window.SwapBuffers()
		glfw.PollEvents()

		inputs.GetInputManager().Notify()
	}
	logger.LOG.Info().Msgf("This is still being used:%v. %v", &myDummyObj.i, myDummyObj.s)
}

func framebufferResizeCallback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

type dummyObj struct {
	s string
	i inputs.InputListener
}

func (d dummyObj) OnKeyAction(a glfw.Action) {
	printToWorld(a, d.s)
}

func printToWorld(a glfw.Action, s string) {
	if a == glfw.Press {
		logger.LOG.Debug().Msg(s + ": Hello World")
	}
	if a == glfw.Release {
		logger.LOG.Debug().Msg(s + ": Goodbye World")
	}
}

func gcLater() {
	timer := time.NewTicker(5 * time.Second)
	<-timer.C
	timer.Stop()
	runtime.GC()
	logger.LOG.Debug().Msg("GC activated")
}

func unsubLater(d *dummyObj) {
	timer := time.NewTicker(10 * time.Second)
	<-timer.C
	timer.Stop()
	inputs.Unsubscribe(glfw.KeyW, weak.Make(&d.i))
}
