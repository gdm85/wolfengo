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
	"math"
)

type Quaternion struct {
	X, Y, Z, W float32
}

func (q Quaternion) length() float32 {
	return float32(math.Sqrt(float64(q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W)))
}

func (q Quaternion) conjugate() Quaternion {
	return Quaternion{-q.X, -q.Y, -q.Z, q.W}
}

func (q Quaternion) mul(v Vector3f) Quaternion {
	w := -q.X*v.X - q.Y*v.Y - q.Z*v.Z
	x := q.W*v.X + q.Y*v.Z - q.Z*v.Y

	y := q.W*v.Y + q.Z*v.X - q.X*v.Z
	z := q.W*v.Z + q.X*v.Y - q.Y*v.X

	return Quaternion{x, y, z, w}
}

func (q Quaternion) mulq(r Quaternion) Quaternion {
	w := q.W*r.W - q.X*r.X - q.Y*r.Y - q.Z*r.Z
	x := q.X*r.W + q.W*r.X + q.Y*r.Z - q.Z*r.Y
	y := q.Y*r.W + q.W*r.Y + q.Z*r.X - q.X*r.Z
	z := q.Z*r.W + q.W*r.Z + q.X*r.Y - q.Y*r.X

	return Quaternion{x, y, z, w}
}
