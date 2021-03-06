package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"./utils"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// 頂点フォーマット
type VertexFormat struct {
	pos   mgl32.Vec3
	color mgl32.Vec4
}

var vertexDatas []VertexFormat
var indexDatas []uint32

// 頂点シェーダプログラム
var vertexShader = `
#version 410

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

in vec3 pv;
in vec4 in_vertexColor;
out vec4 out_vertexColor;

void main() {
	gl_Position = projection * camera * model * vec4(pv, 1);
	out_vertexColor = in_vertexColor;
}
` + "\x00"

// フラグメントシェーダプログラム
var fragmentShader = `
#version 410

in vec4 out_vertexColor;
out vec4 fc;

void main() {
	fc = out_vertexColor;
}
` + "\x00"

// 初期化関数
func init() {
	// 必ずメインスレッドで呼ぶ必要がある
	runtime.LockOSThread()
}

// エントリーポイント
func main() {
	// GLの初期化
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate() // 終了時呼び出し

	// GLFWの設定
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Window作成
	window, err := glfw.CreateWindow(1280, 960, "3D Model Viewer", nil, nil)
	if err != nil {
		panic(err)
	}

	// カレントコンテキスト作成
	window.MakeContextCurrent()

	// GLの初期化
	if err := gl.Init(); err != nil {
		panic(err)
	}
	fmt.Println("OpenGL version", gl.GoStr(gl.GetString(gl.VERSION)))

	// シェーダプログラム作成
	program, err := utils.CreateShaderProgram(vertexShader, fragmentShader)
	if err != nil {
		panic(err)
	}
	gl.UseProgram(program)
	gl.BindFragDataLocation(program, 0, gl.Str("fc\x00"))

	loadModel()

	// 頂点情報作成
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	defer gl.BindVertexArray(0)

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	dataSize := len(vertexDatas) * int(unsafe.Sizeof(vertexDatas[0]))
	gl.BufferData(gl.ARRAY_BUFFER, dataSize, gl.Ptr(vertexDatas), gl.STATIC_DRAW)

	var ibo uint32
	gl.GenBuffers(1, &ibo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	dataSize = len(indexDatas) * int(unsafe.Sizeof(indexDatas[0]))
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, dataSize, gl.Ptr(indexDatas), gl.STATIC_DRAW)

	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("pv\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 3, gl.FLOAT, false, 4*7, gl.PtrOffset(0))

	vertAttrib = uint32(gl.GetAttribLocation(program, gl.Str("in_vertexColor\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointer(vertAttrib, 4, gl.FLOAT, false, 4*7, gl.PtrOffset(3*4))

	// 基本設定
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.ClearColor(0.0, 0.0, 1.0, 1.0)

	angle := 0.0
	previousTime := glfw.GetTime()

	// メインループ
	for !window.ShouldClose() {

		// 画面クリア
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// 使用するシェーダを洗濯
		gl.UseProgram(program)

		// 各行列設定
		projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(1280)/960, 1.0, 10000.0)
		projectionUniform := gl.GetUniformLocation(program, gl.Str("projection\x00"))
		gl.UniformMatrix4fv(projectionUniform, 1, false, &projection[0])

		camera := mgl32.LookAtV(mgl32.Vec3{0, 10, 10}, mgl32.Vec3{0, 3, 0}, mgl32.Vec3{0, 1, 0})
		cameraUniform := gl.GetUniformLocation(program, gl.Str("camera\x00"))
		gl.UniformMatrix4fv(cameraUniform, 1, false, &camera[0])

		time := glfw.GetTime()
		elapsed := time - previousTime
		previousTime = time

		angle += elapsed
		model := mgl32.HomogRotate3D(float32(angle), mgl32.Vec3{0, 1, 0})
		modelUniform := gl.GetUniformLocation(program, gl.Str("model\x00"))
		gl.UniformMatrix4fv(modelUniform, 1, false, &model[0])

		// バッファをバインド
		gl.BindVertexArray(vao)

		// 描画
		gl.DrawElements(gl.TRIANGLES, int32(len(vertexDatas)), gl.UNSIGNED_INT, gl.PtrOffset(0))

		window.SwapBuffers()
		glfw.PollEvents()
	}
}

func loadModel() {
	fmt.Println("test")

	file, err := os.Open(`./gopher.obj`)
	if err != nil {
		// Openエラー処理
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	for i := 1; sc.Scan(); i++ {
		if err := sc.Err(); err != nil {
			// エラー処理
			break
		}

		// 頂点座標読み込み
		if strings.Contains(sc.Text(), "v ") {
			var vertexData VertexFormat

			fmt.Sscanf(sc.Text(), "v %f %f %f", &vertexData.pos[0], &vertexData.pos[1], &vertexData.pos[2])
			vertexData.color[0] = 1.0
			vertexData.color[1] = 1.0
			vertexData.color[2] = 1.0
			vertexData.color[3] = 1.0

			vertexDatas = append(vertexDatas, vertexData)
		}

		// インデックスバッファ読み込み
		if strings.Contains(sc.Text(), "f ") {
			var a, b, c uint32
			fmt.Sscanf(sc.Text(), "f %d %d %d", &a, &b, &c)

			indexDatas = append(indexDatas, a, b, c)
		}
	}
}
