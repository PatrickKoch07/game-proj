package main

import (
	"runtime"
	"time"
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/inputs"
	"github.com/PatrickKoch07/game-proj/internal/logger"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/rs/zerolog"
)

var TARGET_FPS int = 60

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

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Focused, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(1280, 960, "Patrick's Game", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}

	window.SetFramebufferSizeCallback(framebufferResizeCallback)
	window.SetFocusCallback(dummyFocusCallback)
	window.SetCursorPosCallback(dummyMousePosCallback)
	window.SetKeyCallback(inputs.InputKeysCallback)
	window.SetMouseButtonCallback(inputs.InputMouseCallback)

	window.Focus()

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

	fpsLogger := logger.LOG.Sample(&zerolog.BasicSampler{N: uint32(TARGET_FPS)})
	var start_frame_time time.Time
	for !window.ShouldClose() {
		start_frame_time = waitFrame(start_frame_time, fpsLogger)
		// window.SetCursorPos(1280.0/2.0, 960.0/2.0)

		// rendering
		gl.ClearColor(0.2, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		// swaping and polling
		window.SwapBuffers()
		glfw.PollEvents()

		inputs.Notify()
	}
	logger.LOG.Info().Msgf("This is still being used:%v. %v", &myDummyObj.i, myDummyObj.s)
}

func framebufferResizeCallback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func waitFrame(t time.Time, fpsLogger zerolog.Logger) time.Time {
	targetFrameDur := time.Duration(int(1.0/float64(TARGET_FPS)*1000.0) * int(time.Millisecond))
	waitTime := max(targetFrameDur-time.Since(t), 1)
	<-time.NewTicker(waitTime).C

	// below just for displaying framerate
	fps := 1.0 / float64(time.Now().UnixMilli()-t.UnixMilli()) * 1000.0
	fpsLogger.Debug().Msgf(
		"Frame started. Last fps: %v (target: %v). Waited for %v%% of frametime",
		int(fps),
		TARGET_FPS,
		int(float64(waitTime)/float64(targetFrameDur)*100.0),
	)

	return time.Now()
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
	timer := time.NewTicker(3 * time.Second)
	<-timer.C
	timer.Stop()
	runtime.GC()
	logger.LOG.Debug().Msg("GC activated")
}

func unsubLater(d *dummyObj) {
	timer := time.NewTicker(5 * time.Second)
	<-timer.C
	timer.Stop()
	inputs.Unsubscribe(glfw.KeyW, weak.Make(&d.i))
}

func dummyMousePosCallback(w *glfw.Window, xpos float64, ypos float64) {
	w.SetCursorPos(0, 0)
	logger.LOG.Debug().Msgf("Mouse is at (%v, %v)", xpos, ypos)
}

func dummyFocusCallback(w *glfw.Window, focused bool) {
	if focused {
		logger.LOG.Debug().Msgf("Window gained focus")
		w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	} else {
		logger.LOG.Debug().Msgf("Window lost focus")
	}
}
