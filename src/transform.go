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

import "fmt"

type Transform struct {
	translation, rotation, scale Vector3f

	game *Game
}

func (t *Transform) String() string {
	return fmt.Sprintf("{trans: %s, rot: %s, scale: %s}", t.translation.String(), t.rotation.String(), t.scale.String())
}

func (g *Game) NewTransform() *Transform {
	return &Transform{scale: Vector3f{1, 1, 1}, game: g}
}

func (t *Transform) getTransformation() Matrix4f {
	var translationMatrix, rotationMatrix, scaleMatrix Matrix4f
	translationMatrix.initTranslation(t.translation.X, t.translation.Y, t.translation.Z)

	rotationMatrix.initRotation(t.rotation.X, t.rotation.Y, t.rotation.Z)

	scaleMatrix.initScale(t.scale.X, t.scale.Y, t.scale.Z)

	return translationMatrix.mul(rotationMatrix.mul(scaleMatrix))
}

func (t *Transform) getProjectedTransformation(c *Camera) Matrix4f {
	var projectionMatrix, cameraRotation, cameraTranslation Matrix4f

	transformationMatrix := t.getTransformation()
	projectionMatrix.initProjection(c.fov, c.width, c.height, c.zNear, c.zFar)

	cameraRotation.initCamera(c.forward, c.up)

	cameraTranslation.initTranslation(-c.pos.X, -c.pos.Y, -c.pos.Z)

	return projectionMatrix.mul(cameraRotation.mul(cameraTranslation.mul(transformationMatrix)))
}
