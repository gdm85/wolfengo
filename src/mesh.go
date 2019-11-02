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

import "github.com/gdm85/wolfengo/src/gl"

type Mesh struct {
	vbo, ibo uint32
	size     int32
}

func NewMesh(vertices []*Vertex, indices []int32, calcNormals bool) Mesh {
	m := Mesh{}
	m.initMeshData()
	m.addVertices(vertices, indices, calcNormals)
	return m
}

func (m Mesh) IsEmpty() bool {
	return m.vbo == 0
}

func (m *Mesh) initMeshData() {
	gl.GenBuffers(1, &m.vbo)
	gl.GenBuffers(1, &m.ibo)
	m.size = 0
}

func (m *Mesh) addVertices(vertices []*Vertex, indices []int32, calcNormals bool) {
	if calcNormals {
		m.calcNormals(vertices, indices)
	}

	m.size = int32(len(indices))
	fb := verticesAsFloats(vertices)

	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(fb)*4, gl.Ptr(fb), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ibo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)
}

func (m Mesh) draw() {
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)
	gl.EnableVertexAttribArray(2)

	if m.vbo == 0 {
		panic("attempt to set array buffer with VBO=0")
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, VertexSize*4, gl.PtrOffset(0))
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, VertexSize*4, gl.PtrOffset(12))
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, VertexSize*4, gl.PtrOffset(20))

	if m.ibo == 0 {
		panic("attempt to set element array buffer with IBO=0")
	}

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ibo)

	if m.size == 0 {
		panic("attempt to draw elements with mesh size = 0")
	}

	gl.DrawElements(gl.TRIANGLES, m.size, gl.UNSIGNED_INT, gl.PtrOffset(0))

	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
	gl.DisableVertexAttribArray(2)
}

func (m Mesh) calcNormals(vertices []*Vertex, indices []int32) {
	for i := 0; i < len(indices); i += 3 {
		i0, i1, i2 := indices[i], indices[i+1], indices[i+2]

		v1 := vertices[i1].pos.sub(vertices[i0].pos)
		v2 := vertices[i2].pos.sub(vertices[i0].pos)

		normal := v1.cross(v2).normalised()

		vertices[i0].normal = vertices[i0].normal.add(normal)
		vertices[i1].normal = vertices[i1].normal.add(normal)
		vertices[i2].normal = vertices[i2].normal.add(normal)
	}

	for _, v := range vertices {
		v.normal = v.normal.normalised()
	}
}
