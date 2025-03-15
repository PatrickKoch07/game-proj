package ui

import (
	"weak"

	"github.com/PatrickKoch07/game-proj/internal/inputs"
	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type mainMenu struct {
	playButton    *button
	exitButton    *button
	inputListener inputs.InputListener
	rendering     bool
	exitRequested bool
}

var mainMenuObj *mainMenu

func GetMainMenu() *mainMenu {
	if mainMenuObj == nil {
		InitMainMenu()
	}
	return mainMenuObj
}

func InitMainMenu() {
	mainMenuObj = new(mainMenu)
	mainMenuObj.rendering = false

	playButton, err := CreateButton(64, 256, 512, 640)
	if err != nil {
		logger.LOG.Error().Err(err)
		mainMenuObj = nil
		return
	}
	mainMenuObj.playButton = playButton
	mainMenuObj.playButton.OnPress = dummyFunc

	exitButton, err := CreateButton(64, 256, 512, 756)
	if err != nil {
		logger.LOG.Error().Err(err)
		mainMenuObj = nil
		return
	}
	mainMenuObj.exitButton = exitButton
	mainMenuObj.exitButton.OnPress = almostExitGame
	mainMenuObj.exitButton.OnRelease = exitGame

	mainMenuObj.inputListener = inputs.InputListener(mainMenuObj)
	ok := inputs.Subscribe(
		glfw.Key(glfw.KeyEscape),
		weak.Make(&mainMenuObj.inputListener),
	)
	if !ok {
		logger.LOG.Debug().Msg("Main menu failed to subscribe to inputs")
	}
}

func (mm *mainMenu) ShouldClose() bool {
	return mm.exitRequested
}

func (mm *mainMenu) OnKeyAction(a glfw.Action) {
	if a != glfw.Press {
		return
	}
	if mm.rendering {
		mm.stopRenderMainMenu()
	} else {
		mm.renderMainMenu()
	}
	mm.rendering = !mm.rendering
}

func (mm *mainMenu) renderMainMenu() {
	sprites.AddToDrawingQueue(weak.Make(mm.playButton.Sprite))
	sprites.AddToDrawingQueue(weak.Make(mm.exitButton.Sprite))
}

func (mm *mainMenu) stopRenderMainMenu() {
	ok := sprites.RemoveFromDrawingQueue(weak.Make(mm.playButton.Sprite))
	if !ok {
		logger.LOG.Error().Msg("Had trouble removing main menu play button")
	}
	ok = sprites.RemoveFromDrawingQueue(weak.Make(mm.exitButton.Sprite))
	if !ok {
		logger.LOG.Error().Msg("Had trouble removing main menu exit button")
	}
}

func exitGame(_, _ float32) {
	GetMainMenu().exitRequested = true
	// pretty sure this should only get called on the main thread...
	// glfw.GetCurrentContext().SetShouldClose(true)
}

func almostExitGame(_, _ float32) {
	logger.LOG.Debug().Msg("This will exit if you let go over the button!!")
}

func dummyFunc(_, _ float32) {
	logger.LOG.Debug().Msg("Dummy button")
}
