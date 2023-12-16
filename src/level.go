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

	"github.com/gdm85/wolfengo/src/gl"
)

const (
	spotWidth              = 1.0
	spotLength             = 1.0
	spotHeight             = 1.0
	numTexExp              = 4
	openDistance           = 1.0
	doorOpenMovementAmount = 0.9
)

var numTextures = uint32(math.Pow(2, numTexExp))

type Level struct {
	mesh                               Mesh
	level                              *Map
	shader                             *Shader
	material                           *Material
	transform                          *Transform
	player                             *Player
	doors                              []*Door
	monsters                           []*Monster
	medkits                            []*Medkit
	medkitsToRemove                    []*Medkit
	exitPoints                         []*Vector3f
	collisionPosStart, collisionPosEnd []*Vector2f

	crosshairMesh    Mesh
	debugBoxMesh     Mesh
	whiteMaterial    *Material
	debugBoxMaterial *Material

	hud  *HUD
	game *Game // parent game
}

var (
	_basicShader      *Shader
	collectionTexture *Texture
)

func getBasicShader() (*Shader, error) {
	if _basicShader == nil {
		var err error
		collectionTexture, err = NewTexture("WolfCollection.png")
		if err != nil {
			return nil, err
		}

		_basicShader, err = NewShader(true)
		if err != nil {
			return nil, err
		}

		err = _basicShader.addProgramFromFile("basicVertex"+shaderVersion+".vs", gl.VERTEX_SHADER)
		if err != nil {
			return nil, err
		}
		err = _basicShader.addProgramFromFile("basicFragment"+shaderVersion+".fs", gl.FRAGMENT_SHADER)
		if err != nil {
			return nil, err
		}
		err = _basicShader.compile()
		if err != nil {
			return nil, err
		}

		err = _basicShader.addUniform("transform")
		if err != nil {
			return nil, err
		}

		err = _basicShader.addUniform("color")
		if err != nil {
			return nil, err
		}
	}
	return _basicShader, nil
}

func (g *Game) NewLevel(levelNum uint) (*Level, error) {
	l := &Level{game: g}

	l.transform = l.game.NewTransform()

	var filePath string
	if levelNum == 1 && cfgLevelFile != "" {
		filePath = cfgLevelFile
	} else {
		filePath = fmt.Sprintf("./maps/level%d.map", levelNum)
	}

	var err error
	l.level, err = NewMap(filePath)
	if err != nil {
		return nil, err
	}

	l.material = NewMaterial(collectionTexture)

	l.shader, err = getBasicShader()
	if err != nil {
		return nil, err
	}

	err = l.generate()
	if err != nil {
		return nil, err
	}

	// some validation
	if l.player == nil {
		return nil, fmt.Errorf("invalid generated level: no player set")
	}

	whiteTex := NewWhiteTexture()
	l.whiteMaterial = NewMaterial(whiteTex)
	l.whiteMaterial.color = Vector3f{1, 1, 1}

	redTex := NewWhiteTexture()
	l.debugBoxMaterial = NewMaterial(redTex)
	l.debugBoxMaterial.color = Vector3f{1, 0.15, 0}

	l.crosshairMesh = buildCrosshairMesh()
	l.debugBoxMesh = buildUnitBoxMesh()

	l.hud, err = loadHUD(levelNum)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (l *Level) free() {
	for _, m := range l.monsters {
		m.free()
	}
}

func (l *Level) openDoors(position Vector3f, tryExitLevel bool) error {
	for _, door := range l.doors {
		if door.transform.translation.sub(position).length() < openDistance {
			door.open()
		}
	}

	if tryExitLevel {
		for _, exitPoint := range l.exitPoints {
			if exitPoint.sub(position).length() < openDistance {
				err := l.game.loadNextLevel()
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (l *Level) damagePlayer(amt int) {
	l.player.damage(amt)
}

func (l *Level) input() error {
	return l.player.input()
}

func (l *Level) update() error {
	updateAudioListener(l.player.camera.pos, l.player.camera.forward, l.player.camera.up)
	l.hud.setHealth(l.player.health)

	for _, door := range l.doors {
		door.update()
	}

	l.player.update()

	for _, medkit := range l.medkits {
		medkit.update()
	}

	for _, monster := range l.monsters {
		err := monster.update()
		if err != nil {
			return err
		}
	}

	if len(l.medkitsToRemove) > 0 {
		newMedkits := make([]*Medkit, 0, len(l.medkits)-len(l.medkitsToRemove))
		for _, m := range l.medkits {
			removed := false
			for _, r := range l.medkitsToRemove {
				if m == r {
					removed = true
					break
				}
			}
			if !removed {
				newMedkits = append(newMedkits, m)
			}
		}
		l.medkits = newMedkits
		l.medkitsToRemove = nil
	}

	return nil
}

func (l *Level) removeMedkit(m *Medkit) {
	l.medkitsToRemove = append(l.medkitsToRemove, m)
}

func (l *Level) render() {
	l.shader.bind()

	l.shader.updateUniforms(l.transform.getProjectedTransformation(l.player.camera), l.material)
	l.mesh.draw()

	for _, door := range l.doors {
		door.render()
	}

	for _, monster := range l.monsters {
		monster.render()
	}

	for _, medkit := range l.medkits {
		medkit.render()
	}

	l.player.render()

	if cfgDebugHitboxes {
		for _, monster := range l.monsters {
			t := &Transform{
				translation: Vector3f{monster.transform.translation.X - _defaultMonster.sizeX, 0, monster.transform.translation.Z - _defaultMonster.sizeX},
				scale:       Vector3f{defaultMonsterSize.X, _defaultMonster.scale, defaultMonsterSize.Y},
				game:        l.game,
			}
			l.shader.updateUniforms(t.getProjectedTransformation(l.player.camera), l.debugBoxMaterial)
			l.debugBoxMesh.drawLines()
		}
	}

	// Crosshair: rendered in NDC space, ignoring depth buffer
	gl.Disable(gl.DEPTH_TEST)
	var identity Matrix4f
	identity.initIdentity()
	l.shader.updateUniforms(identity, l.whiteMaterial)
	l.crosshairMesh.drawLines()
	gl.Enable(gl.DEPTH_TEST)
}

func (l *Level) renderHUD() {
	l.shader.bind()
	l.hud.render(l.shader, l.player.camera.width, l.player.camera.height)
}

func rectCollide(oldPos, newPos, size1, pos2, size2 Vector2f) (result Vector2f) {
	if newPos.X+size1.X < pos2.X ||
		newPos.X-size1.X > pos2.X+size2.X*size2.X ||
		oldPos.Y+size1.Y < pos2.Y ||
		oldPos.Y-size1.Y > pos2.Y+size2.Y*size2.Y {
		result.X = 1
	}

	if oldPos.X+size1.X < pos2.X ||
		oldPos.X-size1.X > pos2.X+size2.X*size2.X ||
		newPos.Y+size1.Y < pos2.Y ||
		newPos.Y-size1.Y > pos2.Y+size2.Y*size2.Y {
		result.Y = 1
	}

	return
}

func (l *Level) checkCollision(oldPos, newPos Vector3f, objectWidth, objectLength float32) Vector3f {
	collisionVector := Vector2f{1, 1}
	movementVector := newPos.sub(oldPos)

	if movementVector.length() > 0 {
		blockSize := Vector2f{spotWidth, spotLength}
		objectSize := Vector2f{objectWidth, objectLength}

		oldPos2 := Vector2f{oldPos.X, oldPos.Z}
		newPos2 := Vector2f{newPos.X, newPos.Z}

		for i := 0; i < l.level.width; i++ {
			for j := 0; j < l.level.height; j++ {
				if l.level.IsEmpty(i, j) {
					collisionVector = collisionVector.mul(rectCollide(oldPos2, newPos2, objectSize, blockSize.mul(Vector2f{float32(i), float32(j)}), blockSize))
				}
			}
		}

		for _, door := range l.doors {
			doorSize := door.getSize()
			doorPos3f := &door.transform.translation
			doorPos2f := Vector2f{doorPos3f.X, doorPos3f.Z}
			collisionVector = collisionVector.mul(rectCollide(oldPos2, newPos2, objectSize, doorPos2f, doorSize))
		}
	}

	return Vector3f{collisionVector.X, 0, collisionVector.Y}
}

func (l *Level) checkIntersections(lineStart, lineEnd Vector2f, hurtMonsters bool) *Vector2f {
	var nearestIntersection *Vector2f

	for i := 0; i < len(l.collisionPosStart); i++ {
		collisionVector := lineIntersect(lineStart, lineEnd, *l.collisionPosStart[i], *l.collisionPosEnd[i])
		nearestIntersection = findNearestVector2f(nearestIntersection, collisionVector, lineStart)
	}

	// doors stop bullets
	for _, door := range l.doors {
		doorPos3f := door.transform.translation
		doorPos2f := Vector2f{doorPos3f.X, doorPos3f.Z}
		collisionVector := lineIntersectRect(lineStart, lineEnd, doorPos2f, door.getSize())

		nearestIntersection = findNearestVector2f(nearestIntersection, collisionVector, lineStart)
	}

	if hurtMonsters {
		var nearestMonsterIntersect *Vector2f
		var nearestMonster *Monster

		for _, monster := range l.monsters {
			monsterPos3f := monster.transform.translation
			// Offset by half the hitbox size so the AABB is centered on the sprite
			monsterPos2f := Vector2f{monsterPos3f.X - _defaultMonster.sizeX, monsterPos3f.Z - _defaultMonster.sizeX}
			collisionVector := lineIntersectRect(lineStart, lineEnd, monsterPos2f, defaultMonsterSize)

			nearestMonsterIntersect = findNearestVector2f(nearestMonsterIntersect, collisionVector, lineStart)

			if collisionVector != nil && nearestMonsterIntersect == collisionVector {
				nearestMonster = monster
			}
		}

		if nearestMonsterIntersect != nil && (nearestIntersection == nil ||
			nearestMonsterIntersect.sub(lineStart).length() < nearestIntersection.sub(lineStart).length()) {
			if nearestMonster != nil {
				nearestMonster.damage(getPlayerDamage())
			}
		}
	}

	return nearestIntersection
}

func findNearestVector2f(a, b *Vector2f, positionRelativeTo Vector2f) *Vector2f {
	if b != nil && (a == nil ||
		a.sub(positionRelativeTo).length() > b.sub(positionRelativeTo).length()) {
		return b
	}

	return a
}

func lineIntersectRect(lineStart, lineEnd, rectPos, rectSize Vector2f) (result *Vector2f) {
	collisionVector := lineIntersect(lineStart, lineEnd, rectPos, Vector2f{rectPos.X + rectSize.X, rectPos.Y})
	result = findNearestVector2f(result, collisionVector, lineStart)

	collisionVector = lineIntersect(lineStart, lineEnd, rectPos, Vector2f{rectPos.X, rectPos.Y + rectSize.Y})
	result = findNearestVector2f(result, collisionVector, lineStart)

	collisionVector = lineIntersect(lineStart, lineEnd, Vector2f{rectPos.X, rectPos.Y + rectSize.Y}, rectPos.add(rectSize))
	result = findNearestVector2f(result, collisionVector, lineStart)

	collisionVector = lineIntersect(lineStart, lineEnd, Vector2f{rectPos.X + rectSize.X, rectPos.Y}, rectPos.add(rectSize))
	result = findNearestVector2f(result, collisionVector, lineStart)

	return
}

func vector2fCross(a, b Vector2f) float32 {
	return a.X*b.Y - a.Y*b.X
}

// http://stackoverflow.com/questions/563198/how-do-you-detect-where-two-line-segments-intersect
func lineIntersect(lineStart1, lineEnd1, lineStart2, lineEnd2 Vector2f) *Vector2f {
	line1 := lineEnd1.sub(lineStart1)
	line2 := lineEnd2.sub(lineStart2)

	cross := vector2fCross(line1, line2)

	if cross == 0 {
		return nil
	}

	distanceBetweenLineStarts := lineStart2.sub(lineStart1)

	a := vector2fCross(distanceBetweenLineStarts, line2) / cross
	b := vector2fCross(distanceBetweenLineStarts, line1) / cross

	if 0.0 < a && a < 1.0 && 0.0 <= b && b <= 1.0 {
		result := lineStart1.add(line1.mulf(a))
		return &result
	}

	return nil
}

// buildCrosshairMesh creates a line-based + crosshair in NDC space.
// The window is 800x600, so 1px = 2/800 NDC in X and 2/600 NDC in Y.
// Arms are ~10px long with a ~3px gap around the centre.
func buildCrosshairMesh() Mesh {
	const (
		hOuter = float32(0.025)  // 10px in X (NDC)
		hInner = float32(0.0075) // 3px gap in X (NDC)
		vOuter = float32(0.033)  // 10px in Y (NDC)
		vInner = float32(0.010)  // 3px gap in Y (NDC)
	)
	verts := []*Vertex{
		{pos: Vector3f{-hOuter, 0, 0}}, // 0 left end
		{pos: Vector3f{-hInner, 0, 0}}, // 1 inner left
		{pos: Vector3f{hInner, 0, 0}},  // 2 inner right
		{pos: Vector3f{hOuter, 0, 0}},  // 3 right end
		{pos: Vector3f{0, vOuter, 0}},  // 4 top end
		{pos: Vector3f{0, vInner, 0}},  // 5 inner top
		{pos: Vector3f{0, -vInner, 0}}, // 6 inner bottom
		{pos: Vector3f{0, -vOuter, 0}}, // 7 bottom end
	}
	// Four line segments: left arm, right arm, top arm, bottom arm
	indices := []int32{0, 1, 2, 3, 4, 5, 6, 7}
	return NewMesh(verts, indices, false)
}

// buildUnitBoxMesh creates a wireframe unit cube (0..1 in XYZ).
// Scale/translate the resulting transform to position hitboxes.
func buildUnitBoxMesh() Mesh {
	verts := []*Vertex{
		{pos: Vector3f{0, 0, 0}}, // 0 bottom
		{pos: Vector3f{1, 0, 0}}, // 1
		{pos: Vector3f{1, 0, 1}}, // 2
		{pos: Vector3f{0, 0, 1}}, // 3
		{pos: Vector3f{0, 1, 0}}, // 4 top
		{pos: Vector3f{1, 1, 0}}, // 5
		{pos: Vector3f{1, 1, 1}}, // 6
		{pos: Vector3f{0, 1, 1}}, // 7
	}
	// Bottom ring, top ring, 4 vertical pillars — 12 edges = 24 indices
	indices := []int32{
		0, 1, 1, 2, 2, 3, 3, 0, // bottom
		4, 5, 5, 6, 6, 7, 7, 4, // top
		0, 4, 1, 5, 2, 6, 3, 7, // verticals
	}
	return NewMesh(verts, indices, false)
}

func (l *Level) generate() error {
	var vertices []*Vertex
	var indices []int32

	for i := 0; i < l.level.width; i++ {
		for j := 0; j < l.level.height; j++ {
			if l.level.IsEmpty(i, j) {
				continue
			}

			err := l.addSpecial(Special(l.level.specials[i][j]), i, j)
			if err != nil {
				return err
			}

			//Generate Floor
			texCoords := l.level.PlaneTexCoords(i, j)
			addFace(&indices, len(vertices), true)
			v, err := addVertices(i, j, 0, true, false, true, texCoords[:])
			if err != nil {
				return err
			}
			vertices = append(vertices, v...)

			//Generate Ceiling
			addFace(&indices, len(vertices), false)
			v, err = addVertices(i, j, 1, true, false, true, texCoords[:])
			if err != nil {
				return err
			}
			vertices = append(vertices, v...)

			//Generate Walls
			texCoords = l.level.WallTexCoords(i, j)

			if l.level.IsEmpty(i, j-1) {
				l.collisionPosStart = append(l.collisionPosStart, &Vector2f{float32(i * spotWidth), float32(j * spotLength)})
				l.collisionPosEnd = append(l.collisionPosEnd, &Vector2f{float32((i + 1) * spotWidth), float32(j * spotLength)})
				addFace(&indices, len(vertices), false)
				v, err := addVertices(i, 0, j, true, true, false, texCoords[:])
				if err != nil {
					return err
				}
				vertices = append(vertices, v...)
			}
			if l.level.IsEmpty(i, j+1) {
				l.collisionPosStart = append(l.collisionPosStart, &Vector2f{float32(i * spotWidth), float32((j + 1) * spotLength)})
				l.collisionPosEnd = append(l.collisionPosEnd, &Vector2f{float32((i + 1) * spotWidth), float32((j + 1) * spotLength)})
				addFace(&indices, len(vertices), true)
				v, err := addVertices(i, 0, j+1, true, true, false, texCoords[:])
				if err != nil {
					return err
				}
				vertices = append(vertices, v...)
			}

			if l.level.IsEmpty(i-1, j) {
				l.collisionPosStart = append(l.collisionPosStart, &Vector2f{float32(i * spotWidth), float32(j * spotLength)})
				l.collisionPosEnd = append(l.collisionPosEnd, &Vector2f{float32(i * spotWidth), float32((j + 1) * spotLength)})

				addFace(&indices, len(vertices), true)
				v, err := addVertices(0, j, i, false, true, true, texCoords[:])
				if err != nil {
					return err
				}
				vertices = append(vertices, v...)
			}

			if l.level.IsEmpty(i+1, j) {
				l.collisionPosStart = append(l.collisionPosStart, &Vector2f{float32((i + 1) * spotWidth), float32(j * spotLength)})
				l.collisionPosEnd = append(l.collisionPosEnd, &Vector2f{float32((i + 1) * spotWidth), float32((j + 1) * spotLength)})
				addFace(&indices, len(vertices), false)
				v, err := addVertices(0, j, i+1, false, true, true, texCoords[:])
				if err != nil {
					return err
				}
				vertices = append(vertices, v...)
			}
		}
	}

	l.mesh = NewMesh(vertices, indices, false)
	return nil
}

func calcTexCoords(value uint8) []float32 {
	texX := uint32(value) / numTextures
	texY := texX % numTexExp
	texX /= numTexExp

	result := make([]float32, 4)

	result[0] = 1.0 - float32(texX)/float32(numTexExp)
	result[1] = result[0] - 1.0/float32(numTexExp)
	result[3] = 1.0 - float32(texY)/float32(numTexExp)
	result[2] = result[3] - 1.0/float32(numTexExp)

	return result
}

func (l *Level) addSpecial(special Special, x, y int) error {
	switch special {
	case Empty:
		return nil
	case DoorSpecial:
		err := l.addDoor(x, y)
		if err != nil {
			return err
		}
	case PlayerA:
		l.player = l.game.NewPlayer(Vector3f{(float32(x) + 0.5) * spotWidth, 0.4375, (float32(y) + 0.5) * spotLength}, defaultPlayer.mesh, defaultGunMaterial)
	case MonsterSpecial:
		monsterTransform := l.game.NewTransform()
		monsterTransform.translation = Vector3f{(float32(x) + 0.5) * spotWidth, 0, (float32(y) + 0.5) * spotLength}
		l.monsters = append(l.monsters, l.game.NewMonster(monsterTransform, _defaultMonster.animations))
	case SmallMedkit:
		l.medkits = append(l.medkits, l.game.NewMedkit(Vector3f{(float32(x) + 0.5) * spotWidth, 0, (float32(y) + 0.5) * spotLength}))
	case ExitSpecial:
		l.exitPoints = append(l.exitPoints, &Vector3f{(float32(x) + 0.5) * spotWidth, 0, (float32(y) + 0.5) * spotLength})
	default:
		panic(fmt.Sprintf("unrecognized blue value: %d", special))
	}

	return nil
}

func addFace(indices *[]int32, _startLocation int, direction bool) {
	startLocation := int32(_startLocation)
	var add []int32
	if direction {
		add = []int32{
			startLocation + 2,
			startLocation + 1,
			startLocation + 0,
			startLocation + 3,
			startLocation + 2,
			startLocation + 0,
		}
	} else {
		add = []int32{
			startLocation + 0,
			startLocation + 1,
			startLocation + 2,
			startLocation + 0,
			startLocation + 2,
			startLocation + 3,
		}
	}

	*indices = append(*indices, add...)
}

func addVertices(_i, _j, _offset int, x, y, z bool, texCoords []float32) (result []*Vertex, err error) {
	i, j, offset := float32(_i), float32(_j), float32(_offset)

	result = make([]*Vertex, 4)
	if x && z {
		result[0] = &Vertex{Vector3f{(i * spotWidth), (offset) * spotHeight, j * spotLength}, Vector2f{texCoords[1], texCoords[3]}, Vector3f{}}
		result[1] = &Vertex{Vector3f{((i + 1) * spotWidth), (offset) * spotHeight, j * spotLength}, Vector2f{texCoords[0], texCoords[3]}, Vector3f{}}
		result[2] = &Vertex{Vector3f{((i + 1) * spotWidth), (offset) * spotHeight, (j + 1) * spotLength}, Vector2f{texCoords[0], texCoords[2]}, Vector3f{}}
		result[3] = &Vertex{Vector3f{(i * spotWidth), (offset) * spotHeight, (j + 1) * spotLength}, Vector2f{texCoords[1], texCoords[2]}, Vector3f{}}
	} else if x && y {
		result[0] = &Vertex{Vector3f{(i * spotWidth), j * spotHeight, offset * spotLength}, Vector2f{texCoords[1], texCoords[3]}, Vector3f{}}
		result[1] = &Vertex{Vector3f{((i + 1) * spotWidth), j * spotHeight, offset * spotLength}, Vector2f{texCoords[0], texCoords[3]}, Vector3f{}}
		result[2] = &Vertex{Vector3f{((i + 1) * spotWidth), (j + 1) * spotHeight, offset * spotLength}, Vector2f{texCoords[0], texCoords[2]}, Vector3f{}}
		result[3] = &Vertex{Vector3f{(i * spotWidth), (j + 1) * spotHeight, offset * spotLength}, Vector2f{texCoords[1], texCoords[2]}, Vector3f{}}
	} else if y && z {
		result[0] = &Vertex{Vector3f{(offset * spotWidth), i * spotHeight, j * spotLength}, Vector2f{texCoords[1], texCoords[3]}, Vector3f{}}
		result[1] = &Vertex{Vector3f{(offset * spotWidth), i * spotHeight, (j + 1) * spotLength}, Vector2f{texCoords[0], texCoords[3]}, Vector3f{}}
		result[2] = &Vertex{Vector3f{(offset * spotWidth), (i + 1) * spotHeight, (j + 1) * spotLength}, Vector2f{texCoords[0], texCoords[2]}, Vector3f{}}
		result[3] = &Vertex{Vector3f{(offset * spotWidth), (i + 1) * spotHeight, j * spotLength}, Vector2f{texCoords[1], texCoords[2]}, Vector3f{}}
	} else {
		err = fmt.Errorf("Invalid plane used in level generator")
		return
	}
	return
}

func (l *Level) addDoor(x, y int) error {
	doorTransform := l.game.NewTransform()

	xDoor := l.level.IsEmpty(x, y-1) && l.level.IsEmpty(x, y+1)
	yDoor := l.level.IsEmpty(x-1, y) && l.level.IsEmpty(x+1, y)

	if xDoor == yDoor {
		return fmt.Errorf("Level Generation has failed! :( You placed a door in an invalid location at %d,%d", x, y)
	}

	var openPosition *Vector3f

	if yDoor {
		doorTransform.translation = Vector3f{float32(x), 0, float32(y) + spotLength/2}
		t := doorTransform.translation.sub(Vector3f{doorOpenMovementAmount, 0.0, 0.0})
		openPosition = &t
	}

	if xDoor {
		doorTransform.translation = Vector3f{float32(x) + spotWidth/2, 0, float32(y)}
		doorTransform.rotation = Vector3f{0, 90, 0}
		t := doorTransform.translation.sub(Vector3f{0.0, 0.0, doorOpenMovementAmount})
		openPosition = &t
	}

	l.doors = append(l.doors, l.game.NewDoor(doorTransform, l.material, *openPosition))
	return nil
}
