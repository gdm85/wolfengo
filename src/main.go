/* WolfenGo - https://github.com/gdm85/wolfengo
Copyright (C) 2016 gdm85

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/
package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

const (
	version        = "0.1.1"
	debugGL        = true         // extended debugging of GL calls
	printFPS       = true         // print FPS count every second
	debugLevelTest = false        // will load 'levelTest.map'
	frameCap       = float64(250) // cap max framerate to this number of FPS
)

var (
	Window *glfw.Window
	G      *Game
	random *rand.Rand
)

func init() {
	// This is needed to arrange that main() runs on main thread.
	runtime.LockOSThread()

	random = rand.New(rand.NewSource(time.Now().Unix()))
}

func fatalError(err error) {
	fmt.Fprintf(os.Stderr, "FATAL: %v\n", err)
	os.Exit(1)
}

func debugCb(
	source uint32,
	gltype uint32,
	id uint32,
	severity uint32,
	length int32,
	message string,
	userParam unsafe.Pointer) {

	msg := fmt.Sprintf("[GL_DEBUG] source %d gltype %d id %d severity %d length %d: %s", source, gltype, id, severity, length, message)
	if severity == gl.DEBUG_SEVERITY_HIGH {
		panic(msg)
	}
	fmt.Fprintln(os.Stderr, msg)
}

func main() {
	fmt.Printf(`WolfenGo v%s, Copyright (C) 2016 gdm85
https://github.com/gdm85/wolfengo
WolfenGo comes with ABSOLUTELY NO WARRANTY.
This is free software, and you are welcome to redistribute it
under GNU/GPLv2 license.`+"\n", version)

	if err := glfw.Init(); err != nil {
		fatalError(err)
	}
	defer glfw.Terminate()

	// create a context and activate it
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.DoubleBuffer, glfw.True)
	var err error
	Window, err = glfw.CreateWindow(800, 600, "WolfenGo", nil, nil)
	if err != nil {
		fatalError(err)
	}
	Window.MakeContextCurrent()

	// gl.Init() should be called after context is current
	if err := gl.Init(); err != nil {
		fatalError(err)
	}

	glfw.SwapInterval(1) // enable vertical refresh

	fmt.Println(gl.GoStr(gl.GetString(gl.VERSION)))

	// temporary workaround until https://github.com/go-gl/gl/issues/40 is addressed
	if debugGL && runtime.GOOS != "darwin" {
		gl.DebugMessageCallback(debugCb, unsafe.Pointer(nil))
		gl.Enable(gl.DEBUG_OUTPUT)
	}

	// init graphics
	gl.ClearColor(0.0, 0.0, 0.0, 0.0)

	gl.FrontFace(gl.CW)
	gl.CullFace(gl.BACK)
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.DEPTH_TEST)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.Enable(gl.DEPTH_CLAMP)
	gl.Enable(gl.TEXTURE_2D)

	isRunning := false

	// load all assets
	_, err = getBasicShader()
	if err != nil {
		fatalError(err)
	}
	err = _defaultMedkit.initMedkit()
	if err != nil {
		fatalError(err)
	}
	err = _defaultMonster.initMonster()
	if err != nil {
		fatalError(err)
	}
	getDoorMesh()
	initPlayer()
	err = initGun()
	if err != nil {
		fatalError(err)
	}

	G, err = NewGame()
	if err != nil {
		fatalError(err)
	}

	// GAME STARTS
	isRunning = true

	var frames uint64
	var frameCounter time.Duration

	frameTime := time.Duration(1000/frameCap) * time.Millisecond
	lastTime := time.Now()
	var unprocessedTime time.Duration

	for isRunning {
		var render bool

		startTime := time.Now()
		passedTime := startTime.Sub(lastTime)
		lastTime = startTime

		unprocessedTime += passedTime
		frameCounter += passedTime

		for unprocessedTime > frameTime {
			render = true

			unprocessedTime -= frameTime
			if Window.ShouldClose() {
				isRunning = false
			}
			G.timeDelta = frameTime.Seconds()

			glfw.PollEvents()
			err := G.input()
			if err != nil {
				fatalError(err)
			}
			err = G.update()
			if err != nil {
				fatalError(err)
			}

			if frameCounter >= time.Second {
				if printFPS {
					fmt.Printf("%d FPS\n", frames)
				}
				frames = 0
				frameCounter -= time.Second
			}
		}

		if render {
			//TODO: Stencil Buffer
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			G.render()
			frames++
			Window.SwapBuffers()
		} else {
			time.Sleep(time.Millisecond)
		}
	}

	Window.Destroy()
}
