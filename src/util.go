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

type VertexArray []*Vertex

func verticesAsFloats(vertices VertexArray) []float32 {
	buffer := make([]float32, 0, len(vertices)*VertexSize)

	for _, v := range vertices {
		buffer = append(buffer, v.pos.X, v.pos.Y, v.pos.Z)
		buffer = append(buffer, v.texCoord.X, v.texCoord.Y)
		buffer = append(buffer, v.normal.X, v.normal.Y, v.normal.Z)
	}

	return buffer
}

func removeEmptyStrings(a []string) []string {
	result := make([]string, 0, len(a))
	for _, s := range a {
		if len(s) != 0 {
			result = append(result, s)
		}
	}
	return result
}
