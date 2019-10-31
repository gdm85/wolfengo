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

import "time"

const (
	doorLength = 1
	doorHeight = 1
	doorWidth  = 0.125
	doorStart  = 0
	timeToOpen = time.Duration(250) * time.Millisecond
	closeDelay = time.Duration(2) * time.Second
)

var _defaultDoorMesh Mesh

type Door struct {
	mesh                                                    Mesh
	material                                                *Material
	transform                                               *Transform
	openPosition, closePosition                             Vector3f
	isOpening                                               bool
	openingStartTime, openTime, closingStartTime, closeTime time.Time

	game *Game
}

func getDoorMesh() Mesh {
	if _defaultDoorMesh.IsEmpty() {
		vertices := []*Vertex{
			&Vertex{Vector3f{doorStart, doorStart, doorStart}, Vector2f{0.5, 1}, Vector3f{}},
			&Vertex{Vector3f{doorStart, doorHeight, doorStart}, Vector2f{0.5, 0.75}, Vector3f{}},
			&Vertex{Vector3f{doorLength, doorHeight, doorStart}, Vector2f{0.75, 0.75}, Vector3f{}},
			&Vertex{Vector3f{doorLength, doorStart, doorStart}, Vector2f{0.75, 1}, Vector3f{}},

			&Vertex{Vector3f{doorStart, doorStart, doorStart}, Vector2f{0.73, 1}, Vector3f{}},
			&Vertex{Vector3f{doorStart, doorHeight, doorStart}, Vector2f{0.73, 0.75}, Vector3f{}},
			&Vertex{Vector3f{doorStart, doorHeight, doorWidth}, Vector2f{0.75, 0.75}, Vector3f{}},
			&Vertex{Vector3f{doorStart, doorStart, doorWidth}, Vector2f{0.75, 1}, Vector3f{}},

			&Vertex{Vector3f{doorStart, doorStart, doorWidth}, Vector2f{0.5, 1}, Vector3f{}},
			&Vertex{Vector3f{doorStart, doorHeight, doorWidth}, Vector2f{0.5, 0.75}, Vector3f{}},
			&Vertex{Vector3f{doorLength, doorHeight, doorWidth}, Vector2f{0.75, 0.75}, Vector3f{}},
			&Vertex{Vector3f{doorLength, doorStart, doorWidth}, Vector2f{0.75, 1}, Vector3f{}},

			&Vertex{Vector3f{doorLength, doorStart, doorStart}, Vector2f{0.73, 1}, Vector3f{}},
			&Vertex{Vector3f{doorLength, doorHeight, doorStart}, Vector2f{0.73, 0.75}, Vector3f{}},
			&Vertex{Vector3f{doorLength, doorHeight, doorWidth}, Vector2f{0.75, 0.75}, Vector3f{}},
			&Vertex{Vector3f{doorLength, doorStart, doorWidth}, Vector2f{0.75, 1}, Vector3f{}},
		}

		indices := []int32{
			0, 1, 2,
			0, 2, 3,
			6, 5, 4,
			7, 6, 4,
			10, 9, 8,
			11, 10, 8,
			12, 13, 14,
			12, 14, 15}

		_defaultDoorMesh = NewMesh(vertices, indices, false)
	}

	return _defaultDoorMesh
}

func (g *Game) NewDoor(transform *Transform, material *Material, openPosition Vector3f) *Door {
	d := Door{}
	d.game = g

	d.mesh = getDoorMesh()

	d.transform, d.material, d.openPosition = transform, material, openPosition
	d.closePosition = d.transform.translation.mulf(1)

	return &d
}

func (d *Door) open() {
	if d.isOpening {
		return
	}

	d.openingStartTime = time.Now()
	d.openTime = d.openingStartTime.Add(timeToOpen)
	d.closingStartTime = d.openTime.Add(closeDelay)
	d.closeTime = d.closingStartTime.Add(timeToOpen)

	d.isOpening = true
}

func getIncrements(now, target time.Time, delta time.Duration) float32 {
	t := float32(now.Sub(target).Nanoseconds())

	return t / float32(delta.Nanoseconds())
}

func vectorLerp(startPos, endPos Vector3f, lerpFactor float32) Vector3f {
	return startPos.add(endPos.sub(startPos).mulf(lerpFactor))
}

func (d *Door) update() {
	if d.isOpening {
		now := time.Now()

		if now.Before(d.openTime) {
			d.transform.translation = vectorLerp(d.closePosition, d.openPosition, getIncrements(now, d.openingStartTime, timeToOpen))
		} else if now.Before(d.closingStartTime) {
			d.transform.translation = d.openPosition
		} else if now.Before(d.closeTime) {
			d.transform.translation = vectorLerp(d.openPosition, d.closePosition, getIncrements(now, d.closingStartTime, timeToOpen))
		} else {
			d.transform.translation = d.closePosition
			d.isOpening = false
		}
	}
}

func (d *Door) render() {
	t := d.transform.getProjectedTransformation(d.game.Camera())
	d.game.level.shader.updateUniforms(t, d.material)
	d.mesh.draw()
}

func (d *Door) getSize() Vector2f {
	if d.transform.rotation.Y == 90 {
		return Vector2f{doorWidth, doorLength}
	}

	return Vector2f{doorLength, doorWidth}
}
