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
	"fmt"
	"math"
)

const VertexSize = 8

type Vector3f struct {
	X, Y, Z float32
}

func (v *Vector3f) String() string {
	return fmt.Sprintf("[X: %.3f, Y: %.3f, Z: %.3f]", v.X, v.Y, v.Z)
}

type Vector2f struct {
	X, Y float32
}

type Vertex struct {
	pos      Vector3f
	texCoord Vector2f
	normal   Vector3f
}

func (v Vector2f) sub(s Vector2f) Vector2f {
	return Vector2f{v.X - s.X, v.Y - s.Y}
}

func (v Vector2f) add(s Vector2f) Vector2f {
	return Vector2f{v.X + s.X, v.Y + s.Y}
}

func (v Vector2f) mulf(f float32) Vector2f {
	return Vector2f{v.X * f, v.Y * f}
}

func (v Vector2f) length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

func (v Vector2f) rotate(angle float32) Vector2f {
	rad := float64(toRadians(angle))

	cos := float32(math.Cos(rad))
	sin := float32(math.Sin(rad))

	return Vector2f{v.X*cos - v.Y*sin, v.X*sin + v.Y*cos}
}

func (v Vector2f) mul(s Vector2f) Vector2f {
	return Vector2f{v.X * s.X, v.Y * s.Y}
}

func (v Vector2f) normalised() Vector2f {
	length := v.length()

	return Vector2f{v.X / length, v.Y / length}
}

func (v Vector3f) sub(s Vector3f) Vector3f {
	return Vector3f{v.X - s.X, v.Y - s.Y, v.Z - s.Z}
}

func (v Vector3f) add(s Vector3f) Vector3f {
	return Vector3f{v.X + s.X, v.Y + s.Y, v.Z + s.Z}
}

func (v Vector3f) mulf(f float32) Vector3f {
	return Vector3f{v.X * f, v.Y * f, v.Z * f}
}

func (v Vector3f) divf(f float32) Vector3f {
	return Vector3f{v.X / f, v.Y / f, v.Z / f}
}

func (v Vector3f) cross(s Vector3f) Vector3f {
	x := v.Y*s.Z - v.Z*s.Y
	y := v.Z*s.X - v.X*s.Z
	z := v.X*s.Y - v.Y*s.X

	return Vector3f{x, y, z}
}

func (v Vector3f) mul(s Vector3f) Vector3f {
	return Vector3f{v.X * s.X, v.Y * s.Y, v.Z * s.Z}
}

func (v Vector3f) length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

func (v Vector3f) normalised() Vector3f {
	length := v.length()

	return Vector3f{v.X / length, v.Y / length, v.Z / length}
}

func toRadians(degrees float32) float32 {
	return (degrees * math.Pi) / 180
}

func AtanAndToDegrees(radians float32) float32 {
	return float32((math.Atan(float64(radians)) * 180) / math.Pi)
}

func (v Vector3f) rotate(angle float32, axis Vector3f) Vector3f {
	radians := float64(toRadians(angle / 2))
	sinHalfAngle := float32(math.Sin(radians))
	cosHalfAngle := float32(math.Cos(radians))

	rX := axis.X * sinHalfAngle
	rY := axis.Y * sinHalfAngle
	rZ := axis.Z * sinHalfAngle
	rW := cosHalfAngle

	rotation := Quaternion{rX, rY, rZ, rW}
	conjugate := rotation.conjugate()

	w := rotation.mul(v).mulq(conjugate)

	return Vector3f{w.X, w.Y, w.Z}
}
