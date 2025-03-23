package text

import (
	"unicode"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/sprites"
)

const fontSize float32 = 1.0
const baseFontSize int = 16

// -1 for maxWidth or maxHeight means no limit.
// Max width and /n will tell the sprites when to jump to the next line
func TextToSprites(
	message string, bottomLeftX, bottomLeftY, scale float32, maxCharWidth int,
) ([]*sprites.Sprite, bool) {
	textSprites := make([]*sprites.Sprite, len(message))
	ok := true
	var rowCount int = 0
	var colCount int = 0
	for i, char := range message {
		if char == '\n' {
			rowCount++
			colCount = 0
			continue
		}

		screenX := bottomLeftX + float32(colCount*baseFontSize)*scale
		screenY := bottomLeftY + float32(rowCount*baseFontSize)*scale
		textSprite := runeToSprite(screenX, screenY, scale, char)
		if textSprite == nil {
			ok = false
		} else {
			textSprites[i] = textSprite
		}
		colCount++
		if colCount > maxCharWidth {
			rowCount++
			colCount = 0
		}
	}
	return textSprites, ok
}

func runeToSprite(screenX, screenY, scale float32, char rune) *sprites.Sprite {
	var runeToCoords = map[rune][2]int{
		'a': {0, 0}, 'b': {1, 0}, 'c': {2, 0}, 'd': {3, 0}, 'e': {4, 0}, 'f': {5, 0}, 'g': {6, 0},
		'h': {0, 1}, 'i': {1, 1}, 'j': {2, 1}, 'k': {3, 1}, 'l': {4, 1}, 'm': {5, 1}, 'n': {6, 1},
		'o': {0, 2}, 'p': {1, 2}, 'q': {2, 2}, 'r': {3, 2}, 's': {4, 2}, 't': {5, 2}, 'u': {6, 2},
		'v': {0, 3}, 'w': {1, 3}, 'x': {2, 3}, 'y': {3, 3}, 'z': {4, 3}, '?': {5, 3}, '!': {6, 3},
		'1': {0, 4}, '2': {1, 4}, '3': {2, 4}, '4': {3, 4}, '5': {4, 4}, '6': {5, 4}, '7': {6, 4},
		'8': {0, 5}, '9': {1, 5}, '0': {2, 5}, '.': {3, 5}, '$': {4, 5}, '-': {5, 5}, ' ': {6, 5},
	}
	char = unicode.ToLower(char)
	// row number is actually the second value, ie. first row has a, b, c, ...
	return makeSprite(screenX, screenY, scale, runeToCoords[char][1], runeToCoords[char][0])
}

func makeSprite(screenX, screenY, scale float32, row, col int) *sprites.Sprite {
	texCoords := [12]float32{
		float32(col) / 7.0, float32(row) / 6.0,
		float32(col) / 7.0, float32(row+1) / 6.0,
		float32(col+1) / 7.0, float32(row+1) / 6.0,

		float32(col) / 7.0, float32(row) / 6.0,
		float32(col+1) / 7.0, float32(row+1) / 6.0,
		float32(col+1) / 7.0, float32(row) / 6.0,
	}

	return createSprite(screenX, screenY, scale, texCoords)
}

func createSprite(screenX, screenY, scale float32, texCoords [12]float32) *sprites.Sprite {
	sprite, err := sprites.CreateSprite(
		&sprites.SpriteInitParams{
			ShaderRelPaths: sprites.ShaderFiles{
				VertexPath:   "textShader.vs",
				FragmentPath: "alphaTextureShader.fs",
			},
			// 6 rows of 7 columns of characters
			TextureRelPath: "ui/font.png",
			TextureCoords:  texCoords,
			ScreenX:        screenX,
			ScreenY:        screenY,
			SpriteOriginX:  0.0,
			SpriteOriginY:  0.0,
			StretchX:       fontSize * scale,
			StretchY:       fontSize * scale,
		},
	)
	if err != nil {
		logger.LOG.Error().Msg("rune failed to be made.")
		logger.LOG.Error().Err(err).Msg("")
		return nil
	}
	return sprite
}
