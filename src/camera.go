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

var yAxis = Vector3f{0, 1, 0}

type Camera struct {
	pos     Vector3f
	forward Vector3f
	up      Vector3f

	mouseSensitivity float32

	// originally static fields in Transform
	zNear, zFar, width, height, fov float32
}

func NewCamera(pos, forward, up Vector3f, mouseSensitivity float32) *Camera {
	c := Camera{}
	c.pos = pos
	c.forward = forward.normalised()
	c.up = up.normalised()
	c.mouseSensitivity = mouseSensitivity
	w, h := Window.GetSize()
	c.fov, c.width, c.height, c.zNear, c.zFar = 70, float32(w), float32(h), 0.01, 1000.0

	return &c
}

func (c *Camera) mouseLook(oldPosition *Vector2f) {
	x, y := Window.GetCursorPos()
	newPosition := Vector2f{float32(x), float32(y)}
	deltaPos := newPosition.sub(*oldPosition)

	rotY := deltaPos.X != 0
	rotX := deltaPos.Y != 0

	if rotY {
		c.rotateY(deltaPos.X * c.mouseSensitivity)
	}
	if rotX {
		c.rotateX(deltaPos.Y * c.mouseSensitivity)
	}

	if rotY || rotX {
		*oldPosition = newPosition
	}
}

func (c *Camera) rotateY(angle float32) {
	Haxis := yAxis.cross(c.forward).normalised()

	c.forward = c.forward.rotate(angle, yAxis).normalised()

	c.up = c.forward.cross(Haxis).normalised()
}

func (c *Camera) rotateX(angle float32) {
	Haxis := yAxis.cross(c.forward).normalised()

	c.forward = c.forward.rotate(angle, Haxis).normalised()

	c.up = c.forward.cross(Haxis).normalised()
}

func (c *Camera) getLeft() Vector3f {
	return c.forward.cross(c.up).normalised()
}

func (c *Camera) getRight() Vector3f {
	return c.up.cross(c.forward).normalised()
}

func (c *Camera) move(dir Vector3f, amt float32) {
	c.pos = c.pos.add(dir.mulf(amt))
}
