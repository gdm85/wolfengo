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
	_ "image/png"
	"os"
)

type Special byte

const (
	UnknownSpecial         Special = 0
	PlayerA                Special = 'A'
	PlayerB                Special = 'B'
	Pistol                 Special = 'P'
	Gun                    Special = 'G'
	Rocket                 Special = 'R'
	Plasma                 Special = 'S'
	Chaingun               Special = 'C'
	PistolAmmo             Special = 'I'
	GunAmmo                Special = 'U'
	RocketAmmo             Special = 'O'
	PlasmaAmmo             Special = 'L'
	BigMedkit              Special = 'M'
	SmallMedkit            Special = 'm'
	LightAmplificatorVisor Special = 'V'
	DoorSpecial            Special = 'd'
	MonsterSpecial         Special = 'e'
	ExitSpecial            Special = 'X'
	Empty                  Special = ' '
)

type wallDef [4]float32

func (wd *wallDef) String() string {
	return fmt.Sprintf("{%.2f, %.2f, %.2f, %.2f}", wd[0], wd[1], wd[2], wd[3])
}

type Map struct {
	wallDefs                []wallDef
	walls, planes, specials [][]byte
	width, height           int
}

type mapError struct {
	fileName string
	err      error
}

func (be mapError) Error() string {
	return fmt.Sprintf("NewMap(%s): %v", be.fileName, be.err)
}

func NewMap(fileName string) (*Map, error) {
	b := Map{}

	err := b.loadMap(fileName)
	if err != nil {
		return nil, mapError{fileName, err}
	}

	return &b, nil
}

func (m *Map) loadMap(fileName string) error {
	f, err := os.Open("./maps/" + fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	my := map[int]wallDef{}
	var maxWallIndex int
	lineNum := 1
	for {
		var wallIndex int
		var coords wallDef
		read, _ := fmt.Fscanf(f, "wall%d {%f,%f,%f,%f}\n", &wallIndex, &coords[0], &coords[1], &coords[2], &coords[3])

		if read == 0 {
			// finished wall declarations
			break
		} else if read != 5 {
			return fmt.Errorf("invalid wall row at line %d (read %d fields)", lineNum, read)
		}

		// add wall definition
		my[wallIndex] = coords

		if wallIndex > maxWallIndex {
			maxWallIndex = wallIndex
		}

		lineNum++
	}

	// append second natural order
	for i := 1; i <= maxWallIndex; i++ {
		m.wallDefs = append(m.wallDefs, my[i])
	}

	// go back 1 byte, as eaten by Fscanf() to peek-read
	_, err = f.Seek(-1, 1)
	if err != nil {
		return err
	}

	// read map size
	var sz uint
	read, _ := fmt.Fscanf(f, "lengthmap       %3d\nMAP:\n", &sz)
	if read != 1 {
		return fmt.Errorf("no valid lengthmap declaration at line %d", lineNum)
	}
	lineNum++
	lineNum++

	m.width, m.height = int(sz), int(sz)

	// now read all map data
	m.walls, err = readMapBlock(f, m.width, m.height, &lineNum)
	if err != nil {
		return err
	}

	_, err = fmt.Fscanf(f, "PLANES:\n")
	if err != nil {
		return fmt.Errorf("could not match PLANES declaration at line %d (%v)", lineNum, err)
	}
	m.planes, err = readMapBlock(f, m.width, m.height, &lineNum)
	if err != nil {
		return err
	}

	_, err = fmt.Fscanf(f, "SPECIALS:\n")
	if err != nil {
		return fmt.Errorf("could not match SPECIALS declaration at line %d (%v)", lineNum, err)
	}
	m.specials, err = readMapBlock(f, m.width, m.height, &lineNum)
	if err != nil {
		return err
	}

	return nil
}

func readMapBlock(f *os.File, w, h int, lineNum *int) ([][]byte, error) {
	block := make([][]byte, w)
	for row := 0; row < h; row++ {
		block[row] = make([]byte, w)
		n, err := f.Read(block[row])
		if err != nil {
			return block, err
		}
		if n != w {
			return block, fmt.Errorf("not enough data at line %d", *lineNum)
		}

		// skip newline
		var nl [1]byte
		_, err = f.Read(nl[:])
		if err != nil {
			return block, err
		}
		if nl[0] != 0xA {
			return block, fmt.Errorf("invalid line termination at line %d (%d)", *lineNum, nl[0])
		}
		(*lineNum)++
	}
	return block, nil
}

func (m *Map) IsEmpty(x, y int) bool {
	return m.specials[x][y] == ' ' && m.walls[x][y] == ' ' && m.planes[x][y] == ' '
}

func (m *Map) WallTexCoords(x, y int) wallDef {
	wdIndex := m.walls[x][y] - 48 - 1
	return m.wallDefs[wdIndex]
}

func (m *Map) PlaneTexCoords(x, y int) wallDef {
	wdIndex := m.planes[x][y] - 48 - 1
	return m.wallDefs[wdIndex]
}
