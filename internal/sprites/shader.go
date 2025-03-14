package sprites

import (
	"runtime"
	"unsafe"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/utils"

	"github.com/go-gl/gl/v4.1-core/gl"
)

/*
	NEED WAY TO STORE ALREADY LOADED TEXTURES AND SHADERS
*/

func SetTransform(
	shaderId uint32, screenX float32, screenY float32,
) {
	gl.UseProgram(shaderId)
	// transform := mgl32.Translate3D(screenX, screenY/float32(screenHeight), screenY)
	trans := [16]float32{
		1.0, 0.0, 0.0, screenX,
		0.0, 1.0, 0.0, screenY,
		0.0, 0.0, 1.0, screenY,
		0.0, 0.0, 0.0, 1.0,
	}
	// logger.LOG.Error().Msgf("%v", trans)
	var uniformName string = "transform"
	gl.UniformMatrix4fv(
		gl.GetUniformLocation(shaderId, utils.StringToUint8(&uniformName)),
		1,
		true,
		&trans[0],
	)
}

func SetScale(shaderId uint32, stretchY float32, stretchX float32) {
	gl.UseProgram(shaderId)
	scale := [16]float32{
		stretchX, 0.0, 0.0, 0.0,
		0.0, stretchY, 0.0, 0.0,
		0.0, 0.0, stretchY, 0.0,
		0.0, 0.0, 0.0, 1.0,
	}
	// logger.LOG.Error().Msgf("%v", scale)
	var uniformName string = "scale"
	gl.UniformMatrix4fv(
		gl.GetUniformLocation(shaderId, utils.StringToUint8(&uniformName)),
		1,
		true,
		&scale[0],
	)
}

func setProjection(shaderId uint32) {
	gl.UseProgram(shaderId)
	proj := [16]float32{
		2.0 / float32(screenWidth), 0.000000, 0.000000, -1.000000,
		0.000000, 2.0 / float32(0.0-screenHeight), 0.000000, 1.000000,
		0.000000, 0.000000, 2.0 / float32(0.0-screenHeight), 1.000000,
		0.000000, 0.000000, 0.000000, 1.000000,
	}
	logger.LOG.Error().Msgf("%v", proj)
	var uniformName string = "projection"
	gl.UniformMatrix4fv(
		gl.GetUniformLocation(shaderId, utils.StringToUint8(&uniformName)),
		1,
		true,
		&proj[0],
	)
}

func DeleteShaders(shaderIds ...uint32) {
	for sid := range shaderIds {
		gl.DeleteProgram(uint32(sid))
	}
}

func DeleteTextures(textureIds ...uint32) {
	numTextures := int32(len(textureIds))
	gl.DeleteTextures(numTextures, &textureIds[0])
}

func MakeShader(
	vShaderFileName, fShaderFileName string,
) (shaderId uint32, ok bool) {
	logger.LOG.Debug().Msg("Creating new shader")

	vertexCode, err := loadShaderCode(vShaderFileName)
	if err != nil {
		return 0, false
	}
	vertexCodes := make([]*uint8, 1)
	vertexCodes[0] = &vertexCode[0]

	fragmentCode, err := loadShaderCode(fShaderFileName)
	if err != nil {
		return 0, false
	}
	fragmentCodes := make([]*uint8, 1)
	fragmentCodes[0] = &fragmentCode[0]

	sV, sF, ok := compileShader(
		&vertexCodes[0],
		int32(len(vertexCode)),
		&fragmentCodes[0],
		int32(len(fragmentCode)),
	)
	if !ok {
		return 0, false
	}

	shaderId, ok = linkShader(sV, sF)
	if !ok {
		return 0, false
	}

	gl.UseProgram(shaderId)
	var uniformName string = "tex"
	gl.Uniform1i(gl.GetUniformLocation(shaderId, utils.StringToUint8(&uniformName)), 0)

	setProjection(shaderId)

	return shaderId, true
}

func compileShader(
	vertexCode **uint8, lengthVCode int32, fragmentCode **uint8, lengthFCode int32,
) (
	shaderVertex uint32, shaderFragment uint32, ok bool,
) {
	ok = true
	p := runtime.Pinner{}
	defer p.Unpin()
	p.Pin(*vertexCode)
	p.Pin(*fragmentCode)

	shaderVertex = gl.CreateShader(gl.VERTEX_SHADER)
	gl.ShaderSource(shaderVertex, 1, vertexCode, &lengthVCode)
	gl.CompileShader(shaderVertex)
	var okay int32
	gl.GetShaderiv(shaderVertex, gl.COMPILE_STATUS, &okay)
	if okay == 0 {
		logger.LOG.Error().Msg("Vertex shader failed to compile")
		log := make([]byte, 1000)
		gl.GetShaderInfoLog(shaderVertex, 1000, nil, &log[0])
		logger.LOG.Error().Msgf("Error:%v", string(log))
		ok = false
	}

	shaderFragment = gl.CreateShader(gl.FRAGMENT_SHADER)
	gl.ShaderSource(shaderFragment, 1, fragmentCode, &lengthFCode)
	gl.CompileShader(shaderFragment)
	gl.GetShaderiv(shaderFragment, gl.COMPILE_STATUS, &okay)
	if okay == 0 {
		logger.LOG.Error().Msg("Fragment shader failed to compile")
		log := make([]byte, 1000)
		gl.GetShaderInfoLog(shaderFragment, 1000, nil, &log[0])
		logger.LOG.Error().Msgf("Error: %v", string(log))
		ok = false
	}

	return shaderVertex, shaderFragment, ok
}

func linkShader(shaderVertex uint32, shaderFragment uint32) (shaderId uint32, ok bool) {
	shaderId = gl.CreateProgram()
	gl.AttachShader(shaderId, shaderVertex)
	defer gl.DeleteShader(shaderVertex)
	gl.AttachShader(shaderId, shaderFragment)
	defer gl.DeleteShader(shaderFragment)
	gl.LinkProgram(shaderId)

	var okay int32
	gl.GetProgramiv(shaderId, gl.LINK_STATUS, &okay)
	if okay == 0 {
		logger.LOG.Error().Msg("Shader failed to link")
		log := make([]byte, 1000)
		gl.GetProgramInfoLog(shaderId, 1000, nil, &log[0])
		logger.LOG.Error().Msgf("Error: %v", string(log))
		return shaderId, false
	}
	return shaderId, true
}

func GenerateTexture(relativePath string) (TextId uint32, xDim int, yDim int, err error) {
	logger.LOG.Debug().Msg("Creating new texture")

	img, err := loadTextures(relativePath)
	if err != nil {
		return 0, 0, 0, err
	}
	p := runtime.Pinner{}
	defer p.Unpin()
	p.Pin(&img.Pix[0])

	gl.GenTextures(1, &TextId)
	gl.BindTexture(gl.TEXTURE_2D, TextId)
	// unbind texture
	defer gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.TextureParameteri(TextId, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(TextId, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(TextId, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(TextId, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(img.Bounds().Dx()),
		int32(img.Bounds().Dy()),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		unsafe.Pointer(&img.Pix[0]),
	)

	return TextId, img.Bounds().Dx(), img.Bounds().Dy(), nil
}

var screenHeight int
var screenWidth int

func InitShaderScreen(sWidth int, sHeight int) {
	screenHeight = sHeight
	screenWidth = sWidth
}
