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

type Game struct {
	level     *Level
	isRunning bool
	levelNum  uint

	// mouse look fields
	oldPosition Vector2f
	mouseLocked bool

	timeDelta float64
}

func NewGame() (*Game, error) {
	g := Game{}
	g.levelNum = 0
	err := g.loadNextLevel()
	if err != nil {
		return nil, err
	}

	return &g, err
}

func (g *Game) input() error {
	return g.level.input()
}

func (g *Game) update() error {
	if g.isRunning {
		return g.level.update()
	}
	return nil
}

func (g *Game) render() {
	if g.isRunning {
		g.level.render()
	}
}

func (g *Game) Camera() *Camera {
	return g.level.player.camera
}

func (g *Game) loadNextLevel() error {
	var err error
	g.levelNum++
	g.level, err = g.NewLevel(g.levelNum)
	if err != nil {
		return err
	}

	g.isRunning = true

	return nil
}

func (g *Game) UnlockMouse() {
	g.mouseLocked = false
}

func (g *Game) LockMouse() {
	x, y := Window.GetCursorPos()
	g.oldPosition = Vector2f{float32(x), float32(y)}

	g.mouseLocked = true
}
