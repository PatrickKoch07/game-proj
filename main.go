package main

import (
	"runtime"
	"time"

	"github.com/PatrickKoch07/game-proj/internal/cursor"
	"github.com/PatrickKoch07/game-proj/internal/inputs"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/scenes"
	"github.com/PatrickKoch07/game-proj/internal/sprites"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const TARGET_FPS float64 = 60.0

func init() {
	logger.LOG.Info().Msg("Init main")
	// for rendering & window
	runtime.LockOSThread()
	initGLFW()
}

func initGLFW() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.Focused, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
}

func main() {
	defer glfw.Terminate()
	window := createWindow()

	GlobalScene := scenes.CreateGlobalScene()
	InputManager := inputs.GetInputManager()
	// Logger to sample fps every second
	for capFPS := setupFramerateCap(); !window.ShouldClose(); capFPS() {
		// clear previous rendering
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.Clear(gl.DEPTH_BUFFER_BIT)
		// draw
		sprites.DrawDrawQueue()
		window.SwapBuffers()

		// deal with inputs
		glfw.PollEvents()
		InputManager.Notify()

		// update objects
		GlobalScene.Update()
	}
}

func createWindow() *glfw.Window {
	logger.LOG.Info().Msg("Creating new window")

	window, err := glfw.CreateWindow(1280, 960, "Patrick's Game", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		panic(err)
	}
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)
	// gl.Viewport(0, 0, 1280, 960)
	gl.Enable(gl.BLEND)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.DEPTH_TEST)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	sprites.SetShaderScreenSize(1280, 960)

	logger.LOG.Info().Msg("Setting window callbacks")
	window.SetFocusCallback(captureMouseFocusCallback)
	window.SetCursorPosCallback(cursor.UpdateMousePosCallback)
	window.SetKeyCallback(inputs.InputKeysCallback)
	window.SetMouseButtonCallback(inputs.InputMouseCallback)

	window.Focus()

	return window
}

func captureMouseFocusCallback(w *glfw.Window, focused bool) {
	if focused {
		logger.LOG.Debug().Msgf("Window gained focus, capturing mouse.")
		w.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	} else {
		logger.LOG.Debug().Msgf("Window lost focus.")
	}
}

func setupFramerateCap() func() {
	var last_frame_start_time time.Time
	// fpsLogger := logger.LOG.Sample(&zerolog.BasicSampler{N: uint32(TARGET_FPS)})

	return func() {
		targetMilliseconds := 1.0 / TARGET_FPS * 1000.0
		targetFrameDur := time.Duration(int(targetMilliseconds) * int(time.Millisecond))
		waitTime := max(targetFrameDur-time.Since(last_frame_start_time), 1)
		<-time.NewTicker(waitTime).C

		// below just for displaying framerate
		if waitProp := float64(waitTime) / float64(targetFrameDur); waitProp < 0.2 {
			fps := 1.0 / float64(time.Now().UnixMilli()-last_frame_start_time.UnixMilli()) * 1000.0
			logger.LOG.Warn().Msgf(
				"Had slow frame. Last fps: %v (target: %v). Waited for %v%% of frametime",
				int(fps),
				TARGET_FPS,
				int(waitProp*100.0),
			)
		}
		// fpsLogger.Debug().Msgf(
		// 	"Frame started. Last fps: %v (target: %v). Waited for %v%% of frametime",
		// 	int(fps),
		// 	TARGET_FPS,
		// 	int(float64(waitTime)/float64(targetFrameDur)*100.0),
		// )

		last_frame_start_time = time.Now()
	}
}
