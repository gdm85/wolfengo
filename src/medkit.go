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

const (
	pickupDistance = 0.75
	healAmount     = 25
)

var (
	_defaultMedkit = Object{
		K:     0.67857142857142857142857142857143,
		scale: 0.25,
	}
)

func init() {
	_defaultMedkit.sizeY = _defaultMedkit.scale
	_defaultMedkit.sizeX = _defaultMedkit.sizeY / (_defaultMedkit.K * 2.5)
	_defaultMedkit.texMinX = -_defaultMedkit.offsetX
	_defaultMedkit.texMaxX = -1 - _defaultMedkit.offsetX
	_defaultMedkit.texMinY = -_defaultMedkit.offsetY
	_defaultMedkit.texMaxY = 1 - _defaultMedkit.offsetY
}

func (m *Object) initMedkit() error {
	vertices := []*Vertex{
		&Vertex{Vector3f{-m.sizeX, m.start, m.start}, Vector2f{m.texMaxX, m.texMaxY}, Vector3f{0, 0, 0}},
		&Vertex{Vector3f{-m.sizeX, m.sizeY, m.start}, Vector2f{m.texMaxX, m.texMinY}, Vector3f{0, 0, 0}},
		&Vertex{Vector3f{m.sizeX, m.sizeY, m.start}, Vector2f{m.texMinX, m.texMinY}, Vector3f{0, 0, 0}},
		&Vertex{Vector3f{m.sizeX, m.start, m.start}, Vector2f{m.texMinX, m.texMaxY}, Vector3f{0, 0, 0}},
	}

	indices := []int32{0, 1, 2, 0, 2, 3}

	m.mesh = NewMesh(vertices, indices, false)

	t, err := NewTexture("MEDIA0.png")
	if err != nil {
		return err
	}

	m.material = NewMaterial(t)
	return nil
}

type Medkit struct {
	transform *Transform
	mesh      Mesh
	game      *Game
}

func (g *Game) NewMedkit(position Vector3f) *Medkit {
	m := Medkit{}
	m.game = g
	m.mesh = _defaultMedkit.mesh
	m.transform = g.NewTransform()
	m.transform.translation = position
	return &m
}

func (m *Medkit) update() {
	directionToCamera := m.game.Camera().pos.sub(m.transform.translation)

	angleToFaceTheCamera := AtanAndToDegrees(directionToCamera.Z / directionToCamera.X)
	if directionToCamera.X < 0 {
		angleToFaceTheCamera += 180
	}
	m.transform.rotation.Y = angleToFaceTheCamera + 90

	if directionToCamera.length() < pickupDistance {
		if m.game.level.player.health < defaultPlayer.maxHealth {
			m.game.level.removeMedkit(m)
			m.game.level.player.damage(-healAmount)
		}
	}
}

func (m *Medkit) render() {
	m.game.level.shader.updateUniforms(m.transform.getProjectedTransformation(m.game.Camera()), _defaultMedkit.material)
	m.mesh.draw()
}
