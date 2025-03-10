package main

import (
	"runtime"
	"time"

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
	window.SetFocusCallback(captureMouseFocusCallback)
	window.SetCursorPosCallback(debugMousePosCallback)
	window.SetKeyCallback(inputs.InputKeysCallback)
	window.SetMouseButtonCallback(inputs.InputMouseCallback)

	window.Focus()

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

		inputs.GetInputManager().Notify()
	}
}

func framebufferResizeCallback(w *glfw.Window, width int, height int) {
	logger.LOG.Error().Msg("This shouldn't be allowed!!!!")
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

func debugMousePosCallback(w *glfw.Window, xpos float64, ypos float64) {
	w.SetCursorPos(0, 0)
	logger.LOG.Debug().Msgf("Mouse is at (%v, %v)", xpos, ypos)
}

func captureMouseFocusCallback(w *glfw.Window, focused bool) {
	if focused {
		logger.LOG.Debug().Msgf("Window gained focus, capturing mouse.")
		w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	} else {
		logger.LOG.Debug().Msgf("Window lost focus.")
	}
}
