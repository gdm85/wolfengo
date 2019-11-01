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

type Object struct {
	K       float32
	scale   float32
	sizeY   float32
	sizeX   float32
	offsetX float32
	offsetY float32
	texMinX float32
	texMaxX float32
	texMinY float32
	texMaxY float32
	start   float32

	size                 float32
	maxHealth            int
	damageMin, damageMax int
	shootDistance        float32
	moveSpeed            float32

	mesh       Mesh
	material   *Material
	animations []*Texture
}
