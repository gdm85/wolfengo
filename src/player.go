/* WolfenGo - https://github.com/gdm85/wolfengo
Copyright (C) 2016~2019 gdm85

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

	"github.com/go-gl/glfw/v3.1/glfw"
)

const (
	gunOffset              = -0.0875
	playerMouseSensitivity = 1.0 / 3.0
)

type Player struct {
	mesh        Mesh
	gunMaterial *Material

	gunTransform   *Transform
	camera         *Camera
	health         int
	movementVector Vector3f

	game *Game
}

var (
	defaultPlayer = Object{
		K:         1.0379746835443037974683544303797,
		scale:     0.0625,
		damageMin: 20,
		damageMax: 60,
		maxHealth: 100,
		moveSpeed: 5.0,

		size:          0.2,
		shootDistance: 1000.0,
	}

	defaultGunMaterial *Material
)

func init() {
	defaultPlayer.sizeY = defaultPlayer.scale
	defaultPlayer.sizeX = defaultPlayer.sizeY / (defaultPlayer.K * 2)
	defaultPlayer.texMinX = -defaultPlayer.offsetX
	defaultPlayer.texMaxX = -1 - defaultPlayer.offsetX
	defaultPlayer.texMinY = -defaultPlayer.offsetY
	defaultPlayer.texMaxY = 1 - defaultPlayer.offsetY
}

func initPlayer() {
	vertices := []*Vertex{
		&Vertex{Vector3f{-defaultPlayer.sizeX, defaultPlayer.start, defaultPlayer.start}, Vector2f{defaultPlayer.texMaxX, defaultPlayer.texMaxY}, Vector3f{}},
		&Vertex{Vector3f{-defaultPlayer.sizeX, defaultPlayer.sizeY, defaultPlayer.start}, Vector2f{defaultPlayer.texMaxX, defaultPlayer.texMinY}, Vector3f{}},
		&Vertex{Vector3f{defaultPlayer.sizeX, defaultPlayer.sizeY, defaultPlayer.start}, Vector2f{defaultPlayer.texMinX, defaultPlayer.texMinY}, Vector3f{}},
		&Vertex{Vector3f{defaultPlayer.sizeX, defaultPlayer.start, defaultPlayer.start}, Vector2f{defaultPlayer.texMinX, defaultPlayer.texMaxY}, Vector3f{}},
	}

	indices := []int32{0, 1, 2, 0, 2, 3}

	defaultPlayer.mesh = NewMesh(vertices, indices, false)
}

func initGun() error {
	t, err := NewTexture("PISGB0.png")
	if err != nil {
		return err
	}
	defaultGunMaterial = NewMaterial(t)
	return nil
}

func (g *Game) NewPlayer(position Vector3f, playerMesh Mesh, gunMaterial *Material) *Player {
	p := Player{}
	p.game = g
	p.mesh = playerMesh
	p.gunMaterial = gunMaterial
	p.camera = NewCamera(position, Vector3f{0, 0, -1}, Vector3f{0, 1, 0}, playerMouseSensitivity)
	p.health = defaultPlayer.maxHealth
	p.gunTransform = g.NewTransform()
	p.gunTransform.translation = Vector3f{7, 0, 7}

	return &p
}

func (p *Player) damage(amt int) {
	p.health -= amt

	// as this function is used to give health too, check for maximum overflow
	if p.health > defaultPlayer.maxHealth {
		p.health = defaultPlayer.maxHealth
	} else if p.health <= 0 {
		p.game.isRunning = false
		fmt.Println("You just died! GAME OVER")
	}
	// this println was in original clone
	fmt.Println("player health =", p.health)
}

func getPlayerDamage() int {
	return rand.Intn(defaultPlayer.damageMax-defaultPlayer.damageMin) + defaultPlayer.damageMin
}

func (p *Player) update() {
	movAmt := defaultPlayer.moveSpeed * float32(p.game.timeDelta)

	p.movementVector.Y = 0
	if p.movementVector.length() > 0 {
		p.movementVector = p.movementVector.normalised()
	}

	oldPos := p.camera.pos
	newPos := oldPos.add(p.movementVector.mulf(movAmt))

	collisionVector := p.game.level.checkCollision(oldPos, newPos, defaultPlayer.size, defaultPlayer.size)
	p.movementVector = p.movementVector.mul(collisionVector)

	if p.movementVector.length() > 0 {
		p.camera.move(p.movementVector, movAmt)
	}

	p.gunTransform.translation = p.camera.pos.add(p.camera.forward.normalised().mulf(0.105))
	p.gunTransform.translation.Y += gunOffset

	directionToCamera := p.camera.pos.sub(p.gunTransform.translation)

	angleToFaceTheCamera := AtanAndToDegrees(directionToCamera.Z / directionToCamera.X)
	if directionToCamera.X < 0 {
		angleToFaceTheCamera += 180
	}
	p.gunTransform.rotation.Y = angleToFaceTheCamera + 90
}

func (p *Player) input() error {
	if Window.GetKey(glfw.KeyE) == glfw.Press {
		err := p.game.level.openDoors(p.camera.pos, true)
		if err != nil {
			return err
		}
	}

	if Window.GetKey(glfw.KeyEscape) == glfw.Press {
		Window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		p.game.UnlockMouse()
	}

	// wait for left mouse click to lock the camera to the mouse
	if Window.GetMouseButton(glfw.MouseButtonLeft) == glfw.Press {
		if !p.game.mouseLocked {
			Window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
			p.game.LockMouse()
		} else {
			// shoot a bullet
			lineStart := Vector2f{p.camera.pos.X, p.camera.pos.Z}
			castDirection := Vector2f{p.camera.forward.X, p.camera.forward.Z}.normalised()
			lineEnd := lineStart.add(castDirection.mulf(defaultPlayer.shootDistance))

			p.game.level.checkIntersections(lineStart, lineEnd, true)
		}
	}

	p.movementVector = Vector3f{0, 0, 0}

	if Window.GetKey(glfw.KeyW) == glfw.Press {
		p.movementVector = p.movementVector.add(p.camera.forward)
	}
	if Window.GetKey(glfw.KeyS) == glfw.Press {
		p.movementVector = p.movementVector.sub(p.camera.forward)
	}
	if Window.GetKey(glfw.KeyA) == glfw.Press {
		p.movementVector = p.movementVector.add(p.camera.getLeft())
	}
	if Window.GetKey(glfw.KeyD) == glfw.Press {
		p.movementVector = p.movementVector.add(p.camera.getRight())
	}
	if Window.GetKey(glfw.KeyQ) == glfw.Press {
		Window.SetShouldClose(true)
	}
	if p.game.mouseLocked {
		p.camera.mouseLook(&p.game.oldPosition)
	}

	return nil
}

func (p *Player) render() {
	p.game.level.shader.updateUniforms(p.gunTransform.getProjectedTransformation(p.camera), p.gunMaterial)
	p.mesh.draw()
}
