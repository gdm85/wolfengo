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
	"math"
)

type Matrix4f [4][4]float32

func (m *Matrix4f) initIdentity() {
	m[0][0] = 1
	m[0][1] = 0
	m[0][2] = 0
	m[0][3] = 0

	m[1][0] = 0
	m[1][1] = 1
	m[1][2] = 0
	m[1][3] = 0

	m[2][0] = 0
	m[2][1] = 0
	m[2][2] = 1
	m[2][3] = 0

	m[3][0] = 0
	m[3][1] = 0
	m[3][2] = 0
	m[3][3] = 1
	return
}

func (m *Matrix4f) initTranslation(x, y, z float32) {
	m[0][0] = 1
	m[0][1] = 0
	m[0][2] = 0
	m[0][3] = x

	m[1][0] = 0
	m[1][1] = 1
	m[1][2] = 0
	m[1][3] = y

	m[2][0] = 0
	m[2][1] = 0
	m[2][2] = 1
	m[2][3] = z

	m[3][0] = 0
	m[3][1] = 0
	m[3][2] = 0
	m[3][3] = 1
}

func (m *Matrix4f) initRotation(_x, _y, _z float32) {
	var rx, ry, rz Matrix4f

	x, y, z := float64(toRadians(_x)), float64(toRadians(_y)), float64(toRadians(_z))

	rz[0][0] = float32(math.Cos(z))
	rz[0][1] = -float32(math.Sin(z))
	rz[0][2] = 0
	rz[0][3] = 0

	rz[1][0] = float32(math.Sin(z))
	rz[1][1] = float32(math.Cos(z))
	rz[1][2] = 0
	rz[1][3] = 0

	rz[2][0] = 0
	rz[2][1] = 0
	rz[2][2] = 1
	rz[2][3] = 0

	rz[3][0] = 0
	rz[3][1] = 0
	rz[3][2] = 0
	rz[3][3] = 1

	rx[0][0] = 1
	rx[0][1] = 0
	rx[0][2] = 0
	rx[0][3] = 0

	rx[1][0] = 0
	rx[1][1] = float32(math.Cos(x))
	rx[1][2] = -float32(math.Sin(x))
	rx[1][3] = 0

	rx[2][0] = 0
	rx[2][1] = float32(math.Sin(x))
	rx[2][2] = float32(math.Cos(x))
	rx[2][3] = 0

	rx[3][0] = 0
	rx[3][1] = 0
	rx[3][2] = 0
	rx[3][3] = 1

	ry[0][0] = float32(math.Cos(y))
	ry[0][1] = 0
	ry[0][2] = -float32(math.Sin(y))
	ry[0][3] = 0

	ry[1][0] = 0
	ry[1][1] = 1
	ry[1][2] = 0
	ry[1][3] = 0

	ry[2][0] = float32(math.Sin(y))
	ry[2][1] = 0
	ry[2][2] = float32(math.Cos(y))
	ry[2][3] = 0

	ry[3][0] = 0
	ry[3][1] = 0
	ry[3][2] = 0
	ry[3][3] = 1

	*m = rz.mul(ry.mul(rx))
}

func (m *Matrix4f) initScale(x, y, z float32) {
	m[0][0] = x
	m[0][1] = 0
	m[0][2] = 0
	m[0][3] = 0

	m[1][0] = 0
	m[1][1] = y
	m[1][2] = 0
	m[1][3] = 0

	m[2][0] = 0
	m[2][1] = 0
	m[2][2] = z
	m[2][3] = 0

	m[3][0] = 0
	m[3][1] = 0
	m[3][2] = 0
	m[3][3] = 1

}

func (m *Matrix4f) initProjection(fov, width, height, zNear, zFar float32) {
	ar := width / height

	tanHalfFOV := float32(math.Tan(float64(toRadians(fov / 2))))
	zRange := zNear - zFar

	m[0][0] = 1.0 / (tanHalfFOV * ar)
	m[0][1] = 0
	m[0][2] = 0
	m[0][3] = 0
	m[1][0] = 0
	m[1][1] = 1 / tanHalfFOV
	m[1][2] = 0
	m[1][3] = 0
	m[2][0] = 0
	m[2][1] = 0
	m[2][2] = (-zNear - zFar) / zRange
	m[2][3] = 2 * zFar * zNear / zRange
	m[3][0] = 0
	m[3][1] = 0
	m[3][2] = 1
	m[3][3] = 0
}

func (m *Matrix4f) initCamera(forward, up Vector3f) {
	f, r := forward.normalised(), up.normalised()
	r = r.cross(f)
	u := f.cross(r)

	m[0][0] = r.X
	m[0][1] = r.Y
	m[0][2] = r.Z
	m[0][3] = 0

	m[1][0] = u.X
	m[1][1] = u.Y
	m[1][2] = u.Z
	m[1][3] = 0

	m[2][0] = f.X
	m[2][1] = f.Y
	m[2][2] = f.Z
	m[2][3] = 0

	m[3][0] = 0
	m[3][1] = 0
	m[3][2] = 0
	m[3][3] = 1
}

func (m *Matrix4f) mul(r Matrix4f) (res Matrix4f) {
	for i := 0; i < len(m); i++ {
		for j := 0; j < len(m[0]); j++ {
			res[i][j] = m[i][0]*r[0][j] +
				m[i][1]*r[1][j] +
				m[i][2]*r[2][j] +
				m[i][3]*r[3][j]
		}
	}
	return
}

func (m *Matrix4f) asArray() (result [16]float32) {
	index := 0
	for i := 0; i < len(m); i++ {
		for j := 0; j < len(m[0]); j++ {
			result[index] = m[i][j]
			index++
		}
	}
	return
}

func (m *Matrix4f) String() string {
	r := "["
	for i := 0; i < len(m); i++ {
		for j := 0; j < len(m[0]); j++ {
			r += fmt.Sprintf("%.3f, ", m[i][j])
		}
		r = r[:len(r)-2] + " | "
	}
	r = r[:len(r)-2] + "]"

	return r
}
