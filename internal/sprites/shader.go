package sprites

import (
	"unsafe"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/utils"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

func SetTransform(
	shaderId uint32, screenX float32, screenY float32,
) {
	deltaXGL, deltaYGL := mgl32.ScreenToGLCoords(
		int(screenX), int(screenY), screenWidth, screenHeight,
	)
	transform := mgl32.Vec3{deltaXGL, deltaYGL, deltaYGL}

	var uniformName string = "transform"
	gl.UniformMatrix4fv(
		gl.GetUniformLocation(shaderId, utils.StringToUint8(&uniformName)),
		1,
		false,
		(*float32)(gl.Ptr(transform)),
	)
}

func MakeShader(
	vShaderFileName, fShaderFileName string,
) (shaderId uint32, ok bool) {
	vCode, err := loadShaderCode(vShaderFileName)
	if err != nil {
		return 0, false
	}
	var vCodeStart *uint8 = &vCode[0]
	fCode, err := loadShaderCode(fShaderFileName)
	if err != nil {
		return 0, false
	}
	var fCodeStart *uint8 = &fCode[0]

	sV, sF, ok := compileShader(&vCodeStart, &fCodeStart)
	if !ok {
		return 0, false
	}

	shaderId, ok = linkShader(sV, sF)
	if !ok {
		return 0, false
	}

	var uniformName string = "projection"
	projection := mgl32.Ortho(0.0, float32(screenWidth), float32(screenHeight), 0.0, -1.0, 1.0)
	gl.UniformMatrix4fv(
		gl.GetUniformLocation(shaderId, utils.StringToUint8(&uniformName)),
		1,
		false,
		(*float32)(gl.Ptr(projection)),
	)

	return shaderId, true
}

func compileShader(
	vertexCode **uint8, fragmentCode **uint8) (shaderVertex uint32, shaderFragment uint32, ok bool,
) {
	logger.LOG.Info().Msg("Compiling shader")
	ok = true

	shaderVertex = gl.CreateShader(gl.VERTEX_SHADER)
	gl.ShaderSource(shaderVertex, 1, vertexCode, nil)
	gl.CompileShader(shaderVertex)
	var okay int32
	gl.GetShaderiv(shaderVertex, gl.COMPILE_STATUS, &okay)
	if okay == 0 {
		logger.LOG.Error().Msg("Shader failed to compile")
		ok = false
	}

	shaderFragment = gl.CreateShader(gl.FRAGMENT_SHADER)
	gl.ShaderSource(shaderFragment, 1, fragmentCode, nil)
	gl.CompileShader(shaderFragment)
	gl.GetShaderiv(shaderFragment, gl.COMPILE_STATUS, &okay)
	if okay == 0 {
		logger.LOG.Error().Msg("Shader failed to compile")
		ok = false
	}

	return shaderVertex, shaderFragment, ok
}

// func CompileShader(vertexCode *string, fragCode *string) (shaderVertex uint32, shaderFragment uint32, ok bool) {
// 	logger.LOG.Info().Msg("Compiling shader")
// 	ok = true

// 	shaderVertex = gl.CreateShader(gl.VERTEX_SHADER)
// 	var vertexCodeBytes *uint8 = utils.StringToUint8(vertexCode)
// 	gl.ShaderSource(shaderVertex, 1, &vertexCodeBytes, nil)
// 	gl.CompileShader(shaderVertex)
// 	var okay int32
// 	gl.GetShaderiv(shaderVertex, gl.COMPILE_STATUS, &okay)
// 	if okay == 0 {
// 		logger.LOG.Error().Msg("Shader failed to compile")
// 		ok = false
// 	}

// 	shaderFragment = gl.CreateShader(gl.FRAGMENT_SHADER)
// 	var fragmentCodeBytes *uint8 = utils.StringToUint8(fragCode)
// 	gl.ShaderSource(shaderFragment, 1, &fragmentCodeBytes, nil)
// 	gl.CompileShader(shaderFragment)
// 	gl.GetShaderiv(shaderFragment, gl.COMPILE_STATUS, &okay)
// 	if okay == 0 {
// 		logger.LOG.Error().Msg("Shader failed to compile")
// 		ok = false
// 	}

// 	return shaderVertex, shaderFragment, ok
// }

func linkShader(shaderVertex uint32, shaderFragment uint32) (shader uint32, ok bool) {
	logger.LOG.Info().Msg("Attaching shader")

	id := gl.CreateProgram()
	gl.AttachShader(id, shaderVertex)
	gl.AttachShader(id, shaderFragment)
	gl.LinkProgram(id)
	var okay int32
	gl.GetProgramiv(id, gl.LINK_STATUS, &okay)
	if okay == 0 {
		logger.LOG.Error().Msg("Shader failed to compile")
		return id, false
	}
	gl.DeleteShader(shaderVertex)
	gl.DeleteShader(shaderFragment)
	return id, true
}

func GenerateTexture(relativePath string) (TextId uint32, err error) {
	img, err := loadTextures(relativePath)
	if err != nil {
		return 0, err
	}

	gl.GenTextures(1, &TextId)
	gl.BindTexture(gl.TEXTURE_2D, TextId)
	// unbind texture
	defer gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(img.Bounds().Dx()),
		int32(img.Bounds().Dy()),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		unsafe.Pointer(&img.Pix),
	)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	return TextId, nil
}

var screenHeight int
var screenWidth int

func InitShaderScreen(sWidth int, sHeight int) {
	screenHeight = sHeight
	screenWidth = sWidth 
}
