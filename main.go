package main

import (
	"fmt"
	"runtime"
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

// 頂点シェーダプログラム
var vertexShader = `
#version 410

in vec3 pv;
in vec4 in_vertexColor;
out vec4 out_vertexColor;

void main() {
	gl_Position = vec4(pv, 1);
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

	// 頂点データ定義
	vertexData := []VertexFormat{
		{mgl32.Vec3{-0.5, -0.5, 0.0}, mgl32.Vec4{1.0, 1.0, 1.0, 1.0}},
		{mgl32.Vec3{0.5, 0.5, 0.0}, mgl32.Vec4{1.0, 0.0, 0.0, 1.0}},
		{mgl32.Vec3{0.5, -0.5, 0.0}, mgl32.Vec4{0.0, 1.0, 0.0, 1.0}},
		{mgl32.Vec3{-0.5, 0.5, 0.0}, mgl32.Vec4{0.0, 0.0, 1.0, 1.0}},
	}

	vertices := []uint32{
		0, 2, 1, 3,
	}

	// 頂点情報作成
	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	defer gl.BindVertexArray(0)

	fmt.Printf("unsafe.Sizeof(vertexData) = %d", unsafe.Sizeof(vertexData))
	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, ((3*4)+(4*4))*4, gl.Ptr(vertexData), gl.STATIC_DRAW)

	var ibo uint32
	gl.GenBuffers(1, &ibo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

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

	// メインループ
	for !window.ShouldClose() {

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		gl.UseProgram(program)

		gl.BindVertexArray(vao)

		gl.DrawElements(gl.TRIANGLE_FAN, 4, gl.UNSIGNED_INT, gl.PtrOffset(0))

		window.SwapBuffers()
		glfw.PollEvents()
	}
}
