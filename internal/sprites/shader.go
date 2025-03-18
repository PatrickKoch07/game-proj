package sprites

import (
	"runtime"
	"unsafe"

	"github.com/PatrickKoch07/game-proj/internal/logger"
	"github.com/PatrickKoch07/game-proj/internal/utils"

	"github.com/go-gl/gl/v4.1-core/gl"
)

var screenHeight int
var screenWidth int
var activeGraphicsObjects *graphicsObjects

func SetShaderScreenSize(sWidth int, sHeight int) {
	screenHeight = sHeight
	screenWidth = sWidth
}

type ShaderFiles struct {
	VertexPath   string
	FragmentPath string
}

func DeleteShaders(shaderFiles ...ShaderFiles) {
	// delete from active objs and the graphics card
	// TODO: LOCK
	activeGraphicsObjs := getActiveGraphicsObjects()
	for _, shaderFile := range shaderFiles {
		shaderId := activeGraphicsObjs.CurrentlyActiveShaders[shaderFile.VertexPath+shaderFile.FragmentPath]
		delete(activeGraphicsObjs.CurrentlyActiveShaders, shaderFile.VertexPath+shaderFile.FragmentPath)
		gl.DeleteProgram(shaderId)
	}
}

func DeleteTextures(relPaths ...string) {
	// delete from active objs and the graphics card
	// TODO: LOCK
	activeGraphicsObjs := getActiveGraphicsObjects()
	for _, relPath := range relPaths {
		textureId := activeGraphicsObjs.CurrentlyActiveTextures[relPath]
		delete(activeGraphicsObjs.CurrentlyActiveTextures, relPath)
		gl.DeleteTextures(1, &textureId.textureId)
	}
}

func DeleteVAO(manyTextureCoords ...[12]float32) {
	// delete from active objs and the graphics card
	// TODO: LOCK
	activeGraphicsObjs := getActiveGraphicsObjects()
	for _, textureCoords := range manyTextureCoords {
		vaoKey := utils.Float32SliceToString(textureCoords[:])
		VAO := activeGraphicsObjs.CurrentlyActiveVAOs[vaoKey]
		delete(activeGraphicsObjs.CurrentlyActiveVAOs, vaoKey)
		gl.DeleteVertexArrays(1, &VAO)
	}
}

type texture struct {
	textureId uint32
	DimX      float32
	DimY      float32
}

func getTexture(
	relativePath string, textureCoords [12]float32,
) (
	texture, bool,
) {
	// get textureId
	tex, ok := getActiveGraphicsObjects().CurrentlyActiveTextures[relativePath]
	if !ok {
		return texture{}, ok
	}
	// VAO should consist of two triangles.
	// First triangle will be the first three 2-D points provided
	tex.DimX *= textureCoords[4] - textureCoords[2]
	tex.DimY *= textureCoords[3] - textureCoords[1]

	return tex, true
}

func getShader(
	shaderFiles ShaderFiles,
) (shaderId uint32, ok bool) {
	vShaderFileName := shaderFiles.VertexPath
	fShaderFileName := shaderFiles.FragmentPath
	shaderId, ok = getActiveGraphicsObjects().CurrentlyActiveShaders[vShaderFileName+fShaderFileName]
	return shaderId, ok
}

func getVAO(textureCoords [12]float32) (uint32, bool) {
	vaoKey := utils.Float32SliceToString(textureCoords[:])
	vao, ok := getActiveGraphicsObjects().CurrentlyActiveVAOs[vaoKey]
	return vao, ok
}

type graphicsObjects struct {
	CurrentlyActiveShaders  map[string]uint32
	CurrentlyActiveTextures map[string]texture
	CurrentlyActiveVAOs     map[string]uint32
}

func initShader() {
	activeGraphicsObjects = new(graphicsObjects)
	activeGraphicsObjects.CurrentlyActiveShaders = make(map[string]uint32)
	// 32 is the max set by openGL
	activeGraphicsObjects.CurrentlyActiveTextures = make(map[string]texture, 32)
	activeGraphicsObjects.CurrentlyActiveVAOs = make(map[string]uint32)
}

func getActiveGraphicsObjects() *graphicsObjects {
	if activeGraphicsObjects == nil {
		initShader()
	}
	return activeGraphicsObjects
}

func setTransform(
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

func setScale(shaderId uint32, stretchX float32, stretchY float32) {
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
	var uniformName string = "projection"
	gl.UniformMatrix4fv(
		gl.GetUniformLocation(shaderId, utils.StringToUint8(&uniformName)),
		1,
		true,
		&proj[0],
	)
}

func MakeVAO(textureCoords [12]float32) uint32 {
	// NOT THREAD SAFE

	logger.LOG.Info().Msg("Initializing sprite VAO & VBO")
	var VAO, VBO uint32
	var spritePosCoords [12]float32 = [12]float32{
		// Bottom left starting position
		0.0, 0.0,
		0.0, 1.0,
		1.0, 1.0,

		0.0, 0.0,
		1.0, 1.0,
		1.0, 0.0,
	}
	var vertexCoords [24]float32
	// Position X Y, Texture X Y
	for i := 0; i < 6; i++ {
		vertexCoords[4*i] = spritePosCoords[2*i]
		vertexCoords[4*i+1] = spritePosCoords[2*i+1]

		vertexCoords[4*i+2] = textureCoords[2*i]
		vertexCoords[4*i+3] = textureCoords[2*i+1]
	}
	gl.GenVertexArrays(1, &VAO)
	gl.GenBuffers(1, &VBO)

	gl.BindVertexArray(VAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)

	gl.BufferData(gl.ARRAY_BUFFER, 24*4, unsafe.Pointer(&vertexCoords[0]), gl.STATIC_DRAW)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, nil)
	gl.EnableVertexAttribArray(0)

	// unbind
	gl.BindVertexArray(0)

	vaoKey := utils.Float32SliceToString(textureCoords[:])
	getActiveGraphicsObjects().CurrentlyActiveVAOs[vaoKey] = VAO
	return VAO
}

func MakeTexture(relativePath string) (texture, error) {
	// NOT THREAD SAFE

	logger.LOG.Debug().Msg("Creating new texture")
	tex := texture{}

	img, err := loadTextures(relativePath)
	if err != nil {
		return texture{}, err
	}
	tex.DimX = float32(img.Bounds().Dx())
	tex.DimY = float32(img.Bounds().Dy())
	p := runtime.Pinner{}
	defer p.Unpin()
	p.Pin(&img.Pix[0])

	gl.GenTextures(1, &tex.textureId)
	gl.BindTexture(gl.TEXTURE_2D, tex.textureId)
	// unbind texture
	defer gl.BindTexture(gl.TEXTURE_2D, 0)

	gl.TextureParameteri(tex.textureId, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(tex.textureId, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(tex.textureId, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(tex.textureId, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

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

	getActiveGraphicsObjects().CurrentlyActiveTextures[relativePath] = tex

	return tex, nil
}

func MakeShader(
	shaderFiles ShaderFiles,
) (shaderId uint32, ok bool) {
	// NOT THREAD SAFE

	vShaderFileName := shaderFiles.VertexPath
	fShaderFileName := shaderFiles.FragmentPath
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

	getActiveGraphicsObjects().CurrentlyActiveShaders[vShaderFileName+fShaderFileName] = shaderId

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
