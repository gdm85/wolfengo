/*
WolfenGo - https://github.com/gdm85/wolfengo
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
	"time"
)

const (
	stateIdle = iota
	stateChase
	stateAttack
	stateDying
	stateDead
)

const (
	offsetFromGround     = 0.0 // -0.075
	movementStopDistance = 1.5
	shootAngle           = 10.0
	attackChance         = 0.5
)

var (
	monsterAnimationFrames = []string{
		"SSWVA1.png",
		"SSWVB1.png",
		"SSWVC1.png",
		"SSWVD1.png",

		"SSWVE0.png",
		"SSWVF0.png",
		"SSWVG0.png",

		"SSWVH0.png",

		"SSWVI0.png",
		"SSWVJ0.png",
		"SSWVK0.png",
		"SSWVL0.png",
		"SSWVM0.png",
	}
	_defaultMonster = Object{K: 1.9310344827586206896551724137931,
		scale: 0.7,

		offsetX: 0.0, // 0.05
		offsetY: 0.0, // 0.01

		moveSpeed: 2.0,

		size: 0.2,

		shootDistance: 1000.0,

		maxHealth: 100,
		damageMin: 5,
		damageMax: 30,
	}
	defaultMonsterSize = Vector2f{_defaultMonster.size, _defaultMonster.size}
)

func init() {
	_defaultMonster.sizeY = _defaultMonster.scale
	_defaultMonster.sizeX = _defaultMonster.sizeY / (_defaultMonster.K * 2)
	_defaultMonster.texMinX = -_defaultMonster.offsetX
	_defaultMonster.texMaxX = -1 - _defaultMonster.offsetX
	_defaultMonster.texMinY = -_defaultMonster.offsetY
	_defaultMonster.texMaxY = 1 - _defaultMonster.offsetY
}

func (m *Object) initMonster() error {
	m.animations = make([]*Texture, len(monsterAnimationFrames))
	for i := 0; i < len(monsterAnimationFrames); i++ {
		var err error
		m.animations[i], err = NewTexture(monsterAnimationFrames[i])
		if err != nil {
			return err
		}
	}

	vertices := []*Vertex{
		&Vertex{Vector3f{-m.sizeX, m.start, m.start}, Vector2f{m.texMaxX, m.texMaxY}, Vector3f{0, 0, 0}},
		&Vertex{Vector3f{-m.sizeX, m.sizeY, m.start}, Vector2f{m.texMaxX, m.texMinY}, Vector3f{0, 0, 0}},
		&Vertex{Vector3f{m.sizeX, m.sizeY, m.start}, Vector2f{m.texMinX, m.texMinY}, Vector3f{0, 0, 0}},
		&Vertex{Vector3f{m.sizeX, m.start, m.start}, Vector2f{m.texMinX, m.texMaxY}, Vector3f{0, 0, 0}},
	}

	indices := []int32{0, 1, 2, 0, 2, 3}

	m.mesh = NewMesh(vertices, indices, false)
	return nil
}

type Monster struct {
	transform  *Transform
	state      int
	canAttack  bool
	canLook    bool
	health     int
	material   *Material
	deathTime  time.Time
	animations []*Texture
	mesh       Mesh

	game *Game
}

func (g *Game) NewMonster(t *Transform, animations []*Texture) *Monster {
	m := Monster{}

	m.mesh = _defaultMonster.mesh
	m.transform = t
	m.game = g
	m.state = stateIdle
	m.health = _defaultMonster.maxHealth
	m.animations = animations
	m.material = NewMaterial(m.animations[0])

	return &m
}

func (m *Monster) damage(amt int) {
	if m.state == stateIdle {
		m.state = stateChase
	}

	m.health -= amt

	if m.health <= 0 {
		m.state = stateDying
	}
}

func getDecimals() float32 {
	now := time.Now()
	ns := float32(now.UnixNano() - now.Unix()*1e9)

	return ns / float32(1e9)
}

func (m *Monster) idleUpdate(orientation Vector3f, distance float32) {
	if getDecimals() < 0.5 {
		m.canLook = true
		m.material.texture = m.animations[0]
	} else {
		if m.canLook {
			lineStart := Vector2f{m.transform.translation.X, m.transform.translation.Z}
			castDirection := Vector2f{orientation.X, orientation.Z}
			lineEnd := lineStart.add(castDirection.mulf(_defaultMonster.shootDistance))

			collisionVector := m.game.level.checkIntersections(lineStart, lineEnd, false)
			playerIntersectVector := Vector2f{m.game.Camera().pos.X, m.game.Camera().pos.Z}

			if collisionVector == nil || playerIntersectVector.sub(lineStart).length() < collisionVector.sub(lineStart).length() {
				m.state = stateChase
			}

			m.canLook = false
		}
	}
}

func (m *Monster) chaseUpdate(orientation Vector3f, distance float32) error {
	timeDecimals := getDecimals()

	var animFrame int
	if timeDecimals < 0.25 {
		animFrame = 0
	} else if timeDecimals < 0.5 {
		animFrame = 1
	} else if timeDecimals < 0.75 {
		animFrame = 2
	} else {
		animFrame = 3
	}
	m.material.texture = m.animations[animFrame]

	if random.Float32() < attackChance*float32(m.game.timeDelta) {
		m.state = stateAttack
	}

	if distance > movementStopDistance {
		moveAmount := _defaultMonster.moveSpeed * float32(m.game.timeDelta)

		oldPos := m.transform.translation
		newPos := m.transform.translation.add(orientation.mulf(moveAmount))

		collisionVector := m.game.level.checkCollision(oldPos, newPos, _defaultMonster.size, _defaultMonster.size)
		movementVector := collisionVector.mul(orientation)

		if movementVector.length() > 0 {
			m.transform.translation = m.transform.translation.add(movementVector.mulf(moveAmount))
		}

		if movementVector.sub(orientation).length() != 0 {
			err := m.game.level.openDoors(m.transform.translation, false)
			if err != nil {
				return err
			}
		}
	} else {
		m.state = stateAttack
	}

	return nil
}

func (m *Monster) attackUpdate(orientation Vector3f, distance float32) {
	timeDecimals := getDecimals()

	if timeDecimals < 0.25 {
		m.material.texture = m.animations[4]
	} else if timeDecimals < 0.5 {
		m.material.texture = m.animations[5]
	} else if timeDecimals < 0.75 {
		m.material.texture = m.animations[6]
		if m.canAttack {
			lineStart := Vector2f{m.transform.translation.X, m.transform.translation.Z}
			castDirection := Vector2f{orientation.X, orientation.Z}.rotate((random.Float32() - 0.5) * shootAngle)
			lineEnd := lineStart.add(castDirection.mulf(_defaultMonster.shootDistance))

			collisionVector := m.game.level.checkIntersections(lineStart, lineEnd, false)

			playerIntersectVector := lineIntersectRect(lineStart, lineEnd, Vector2f{m.game.Camera().pos.X, m.game.Camera().pos.Z}, Vector2f{defaultPlayer.size, defaultPlayer.size})
			if playerIntersectVector != nil && (collisionVector == nil || playerIntersectVector.sub(lineStart).length() < collisionVector.sub(lineStart).length()) {
				m.game.level.damagePlayer(_defaultMonster.damageMin + random.Intn(_defaultMonster.damageMax-_defaultMonster.damageMin))
			}

			m.canAttack = false
		}
	} else {
		m.material.texture = m.animations[5]
		m.state = stateChase
		m.canAttack = true
	}
}

const (
	time1 = time.Duration(100) * time.Millisecond
	time2 = time.Duration(300) * time.Millisecond
	time3 = time.Duration(450) * time.Millisecond
	time4 = time.Duration(600) * time.Millisecond
)

func (m *Monster) dyingUpdate(orientation Vector3f, distance float32) {
	now := time.Now()

	if m.deathTime.IsZero() {
		m.deathTime = now
	}

	if now.Before(m.deathTime.Add(time1)) {
		m.material.texture = m.animations[8]
		m.transform.scale = Vector3f{1, 0.96428571428571428571428571428571, 1}
	} else if now.Before(m.deathTime.Add(time2)) {
		m.material.texture = m.animations[9]
		m.transform.scale = Vector3f{1.7, 0.9, 1.0}
	} else if now.Before(m.deathTime.Add(time3)) {
		m.material.texture = m.animations[10]
		m.transform.scale = Vector3f{1.7, 0.9, 1.0}
	} else if now.Before(m.deathTime.Add(time4)) {
		m.material.texture = m.animations[11]
		m.transform.scale = Vector3f{1.7, 0.5, 1.0}
	} else {
		m.state = stateDead
	}
}

func (m *Monster) deadUpdate(orientation Vector3f, distance float32) {
	m.material.texture = m.animations[12]
	m.transform.scale = Vector3f{1.7586206896551724137931034482759, 0.28571428571428571428571428571429, 1}
}

func (m *Monster) alignWithGround() {
	m.transform.translation.Y = offsetFromGround
}

func (m *Monster) faceCamera(directionToCamera Vector3f) {
	angleToFaceTheCamera := AtanAndToDegrees(directionToCamera.Z / directionToCamera.X)

	if directionToCamera.X < 0 {
		angleToFaceTheCamera += 180
	}

	m.transform.rotation.Y = angleToFaceTheCamera + 90
}

func (m *Monster) update() error {
	directionToCamera := m.game.Camera().pos.sub(m.transform.translation)

	distance := directionToCamera.length()

	orientation := directionToCamera.divf(distance)

	m.alignWithGround()
	m.faceCamera(orientation)

	switch m.state {
	case stateIdle:
		m.idleUpdate(orientation, distance)
	case stateChase:
		err := m.chaseUpdate(orientation, distance)
		if err != nil {
			return err
		}
	case stateAttack:
		m.attackUpdate(orientation, distance)
	case stateDying:
		m.dyingUpdate(orientation, distance)
	case stateDead:
		m.deadUpdate(orientation, distance)
	default:
		panic("unexpected monster state")
	}

	return nil
}

func (m *Monster) render() {
	m.game.level.shader.updateUniforms(m.transform.getProjectedTransformation(m.game.Camera()), m.material)
	m.mesh.draw()
}
