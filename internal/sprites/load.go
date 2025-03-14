package sprites

import (
	"image"
	"image/draw"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/PatrickKoch07/game-proj/internal/logger"
)

func loadShaderCode(fileName string) ([]byte, error) {
	file, err := os.Open(filepath.Join(".", "assets", "shaders", fileName))
	if err != nil {
		logger.LOG.Fatal().Msgf("Opening shader file %v failed", fileName)
		return make([]byte, 0), err
	}
	defer file.Close()

	// seems like a safe limit given our current shaders
	data := make([]byte, 1000)
	count, err := file.Read(data)
	if err != nil {
		logger.LOG.Fatal().Msgf("Loading from shader file %v failed", fileName)
		return make([]byte, 0), err
	}
	if count > 900 {
		logger.LOG.Warn().Msgf("Shader file (%v) is getting long. (%v/1000 char)", fileName, count)
	}

	shaderCode := data[:count]

	return shaderCode, nil
}

func loadTextures(relativePath string) (*image.RGBA, error) {
	fileReader, err := os.Open(filepath.Join(".", "assets", "sprites", relativePath))
	if err != nil {
		logger.LOG.Fatal().Msgf("Opening texture file %v failed", relativePath)
		return nil, err
	}
	defer fileReader.Close()

	raw_image, _, err := image.Decode(fileReader)
	if err != nil {
		logger.LOG.Fatal().Msgf("Opening texture file %v failed", relativePath)
		return nil, err
	}
	b := raw_image.Bounds()

	if sz := (b.Max.Y - b.Min.Y) * (b.Max.X - b.Min.X); sz > (screenWidth * screenHeight) {
		logger.LOG.Fatal().Msgf("File to load has too much data: %v bytes", sz)
		return nil, err
	}

	rgba_image := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(rgba_image, rgba_image.Bounds(), raw_image, b.Min, draw.Src)
	return rgba_image, nil
}
